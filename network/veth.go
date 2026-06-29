package network

import (
	"fmt"
	"os/exec"
)

// SetupVeth generates virtual network pairs and connects container interfaces
func SetupVeth(containerPID int) error {
	fmt.Printf("[Network] Setting up virtual cable for container PID %d\n", containerPID)

	vethHost := fmt.Sprintf("veth_%d", containerPID)
	vethChild := fmt.Sprintf("vethc_%d", containerPID)

	// Clean up leftover interfaces if any
	exec.Command("ip", "link", "del", vethHost).Run()

	// Create veth pair: vethHost <-> vethChild
	err := exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethChild).Run()
	if err != nil {
		return fmt.Errorf("failed to create veth pair (%s): %v", vethHost, err)
	}

	// Attach vethHost to the bridge
	err = exec.Command("ip", "link", "set", vethHost, "master", "md-br0").Run()
	if err != nil {
		exec.Command("ip", "link", "del", vethHost).Run()
		return fmt.Errorf("failed to attach %s to bridge: %v", vethHost, err)
	}

	// Bring vethHost up
	err = exec.Command("ip", "link", "set", vethHost, "up").Run()
	if err != nil {
		exec.Command("ip", "link", "del", vethHost).Run()
		return fmt.Errorf("failed to bring %s up: %v", vethHost, err)
	}

	// Move vethChild into the container's network namespace
	pidStr := fmt.Sprintf("%d", containerPID)
	err = exec.Command("ip", "link", "set", vethChild, "netns", pidStr).Run()
	if err != nil {
		exec.Command("ip", "link", "del", vethHost).Run()
		return fmt.Errorf("failed to move %s into container netns: %v", vethChild, err)
	}

	// Bring lo up inside container
	exec.Command("nsenter", "-t", pidStr, "-n", "ip", "link", "set", "lo", "up").Run()

	// Configure vethChild inside the container using nsenter: rename to eth0 and assign IP
	exec.Command("nsenter", "-t", pidStr, "-n", "ip", "link", "set", vethChild, "name", "eth0").Run()

	// Calculate container IP derived from PID or fallback to 172.20.0.2
	containerIP := fmt.Sprintf("172.20.0.%d", 2+(containerPID%250))
	err = exec.Command(
		"nsenter", "-t", pidStr, "-n",
		"ip", "addr", "add", containerIP+"/24", "dev", "eth0",
	).Run()
	if err != nil {
		// Fallback to vethChild if rename wasn't successful
		err = exec.Command(
			"nsenter", "-t", pidStr, "-n",
			"ip", "addr", "add", containerIP+"/24", "dev", vethChild,
		).Run()
		if err != nil {
			return fmt.Errorf("failed to assign IP inside container: %v", err)
		}
		exec.Command("nsenter", "-t", pidStr, "-n", "ip", "link", "set", vethChild, "up").Run()
	} else {
		exec.Command("nsenter", "-t", pidStr, "-n", "ip", "link", "set", "eth0", "up").Run()
	}

	// Set the default gateway inside the container
	err = exec.Command(
		"nsenter", "-t", pidStr, "-n",
		"ip", "route", "add", "default", "via", "172.20.0.1",
	).Run()
	if err != nil {
		return fmt.Errorf("failed to add default route in container: %v", err)
	}

	fmt.Printf("[Network] Virtual cable connected successfully with IP %s!\n", containerIP)
	return nil
}

// CleanupVeth removes the host-side veth interface for a container PID
func CleanupVethForPID(containerPID int) {
	vethHost := fmt.Sprintf("veth_%d", containerPID)
	fmt.Printf("[Network] Removing virtual cable %s...\n", vethHost)
	exec.Command("ip", "link", "del", vethHost).Run()
}

// CleanupVeth removes generic host-side veth interfaces
func CleanupVeth() {
	fmt.Println("[Network] Removing virtual cable...")
	exec.Command("ip", "link", "del", "veth_host").Run()
}
