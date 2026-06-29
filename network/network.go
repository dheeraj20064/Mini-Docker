package network

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// DetectDefaultInterface finds the active host network interface with default route
func DetectDefaultInterface() string {
	cmd := exec.Command("sh", "-c", "ip route show default | awk '/default/ {print $5}'")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		iface := strings.TrimSpace(out.String())
		if iface != "" {
			return iface
		}
	}
	return "eth0"
}

func SetupNetwork(containerPID int, hostInterface string) error {
	fmt.Println("[Network] Starting network setup...")

	if hostInterface == "" || hostInterface == "eth0" {
		hostInterface = DetectDefaultInterface()
	}

	if err := CreateBridge(); err != nil {
		return fmt.Errorf("bridge setup failed: %v", err)
	}

	if err := SetupNAT(hostInterface); err != nil {
		return fmt.Errorf("NAT setup failed: %v", err)
	}

	if err := SetupVeth(containerPID); err != nil {
		return fmt.Errorf("veth setup failed: %v", err)
	}

	fmt.Println("[Network] Container has internet access!")
	return nil
}

func CleanupContainerNetwork(containerPID int, hostInterface string) {
	fmt.Println("[Network] Cleaning up container network...")
	CleanupVethForPID(containerPID)
}

func CleanupNetwork(hostInterface string) {
	fmt.Println("[Network] Cleaning up network...")
	CleanupVeth()
	CleanupNAT(hostInterface)
}