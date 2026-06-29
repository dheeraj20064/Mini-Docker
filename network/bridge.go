package network

import (
	"fmt"
	"os/exec"
)

// CreateBridge provisions the host-side virtual software switch (md-br0)
func CreateBridge() error {
	fmt.Println("[Network] Creating bridge md-br0...")
	
	// Create the bridge interface
	err := exec.Command("ip", "link", "add", "name", "md-br0", "type", "bridge").Run()
	if err != nil {
		// If the bridge already exists, we can ignore the error
		fmt.Println("[Network] Bridge md-br0 may already exist, skipping creation.")
	}

	// Assign IP address to the bridge
	err = exec.Command("ip", "addr", "add", "172.20.0.1/24", "dev", "md-br0").Run()
	if err != nil {
		fmt.Println("[Network] IP address for md-br0 may already exist, skipping assignment.")
	}

	// Bring the bridge interface up
	err = exec.Command("ip", "link", "set", "md-br0", "up").Run()
	if err != nil {
		return fmt.Errorf("failed to set up bridge md-br0: %v", err)
	}

	fmt.Println("[Network] Bridge is ready!")
	return nil
}

// DeleteBridge removes the bridge interface from the host
func DeleteBridge() {
	fmt.Println("[Network] Removing bridge md-br0...")
	exec.Command("ip", "link", "delete", "md-br0").Run()
}
