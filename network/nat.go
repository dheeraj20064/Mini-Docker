package network

import (
    "fmt"
    "os/exec"
)
func SetupNAT(hostInterface string) error {
    fmt.Printf("[Network] Setting up internet sharing via %s...\n", hostInterface)
	err := exec.Command(
        "sysctl", "-w", "net.ipv4.ip_forward=1",
    ).Run()
    if err != nil {
        return fmt.Errorf("failed to enable IP forwarding: %v", err)
    }

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
	func CleanupNAT(hostInterface string) {
    exec.Command(
        "iptables", "-t", "nat",
        "-D", "POSTROUTING",
        "-s", "172.20.0.0/24",
        "-o", hostInterface,
        "-j", "MASQUERADE",
    ).Run()
    fmt.Println("[Network] NAT rules removed.")
}
}