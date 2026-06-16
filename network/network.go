package network

import "fmt"

func SetupNetwork(containerPID int, hostInterface string) error {
    fmt.Println("[Network] Starting network setup...")


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


func CleanupNetwork(hostInterface string) {
    fmt.Println("[Network] Cleaning up network...")
    CleanupVeth()
    CleanupNAT(hostInterface)

}