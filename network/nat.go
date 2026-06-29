package network

import (
	"fmt"
	"os/exec"
)

// SetupNAT deploys masqueraded iptables network sharing rules
func SetupNAT(hostInterface string) error {
	fmt.Printf("[Network] Setting up internet sharing via %s...\n", hostInterface)

	// Enable IP forwarding in the kernel
	err := exec.Command(
		"sysctl", "-w", "net.ipv4.ip_forward=1",
	).Run()
	if err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %v", err)
	}

	// Add NAT masquerade rule for the container subnet
	err = exec.Command(
		"iptables", "-t", "nat",
		"-A", "POSTROUTING",
		"-s", "172.20.0.0/24",
		"-o", hostInterface,
		"-j", "MASQUERADE",
	).Run()
	if err != nil {
		return fmt.Errorf("failed to add NAT rule: %v", err)
	}

	// Allow forwarding from bridge to host interface
	err = exec.Command(
		"iptables",
		"-A", "FORWARD",
		"-i", "md-br0",
		"-o", hostInterface,
		"-j", "ACCEPT",
	).Run()
	if err != nil {
		return fmt.Errorf("failed to add FORWARD rule (outbound): %v", err)
	}

	// Allow forwarding from host interface to bridge
	err = exec.Command(
		"iptables",
		"-A", "FORWARD",
		"-i", hostInterface,
		"-o", "md-br0",
		"-j", "ACCEPT",
	).Run()
	if err != nil {
		return fmt.Errorf("failed to add FORWARD rule (inbound): %v", err)
	}

	fmt.Println("[Network] Internet sharing is active!")
	return nil
}

// CleanupNAT removes the NAT rules from the system
func CleanupNAT(hostInterface string) {
	fmt.Println("[Network] Removing NAT rules...")
	exec.Command(
		"iptables", "-t", "nat",
		"-D", "POSTROUTING",
		"-s", "172.20.0.0/24",
		"-o", hostInterface,
		"-j", "MASQUERADE",
	).Run()
}
