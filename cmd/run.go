package cmd

import (
	"crypto/rand"
	"fmt"
	"mini-docker/container"
	"mini-docker/network"
)

// generateID creates unique container ID
func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("minidocker-%x", b)
}

func Run(config RunConfig) error {

	// Generate unique ID every time!
	containerID := generateID()
	fmt.Println("Container ID:", containerID)

	fmt.Println("=== Mini Docker ===")
	fmt.Printf("Image:   %s\n", config.Image)
	fmt.Printf("Command: %s\n", config.Command)
	fmt.Printf("Memory:  %dMB\n", config.MemoryMB)
	fmt.Printf("CPU:     %d%%\n", config.CPUPercent)
	fmt.Println("==================")

	// Placeholder for Ameen
	rootfs := ""

	// Step 1: Setup cgroup
	err := container.SetupCgroup(container.CgroupConfig{
		ContainerID: containerID,
		MemoryMB:    config.MemoryMB,
		CPUPercent:  config.CPUPercent,
	})
	if err != nil {
		return fmt.Errorf("cgroup error: %v", err)
	}

	// Step 2: Start isolated container
	pid, err := container.StartContainer(container.ContainerConfig{
		ID:       containerID,
		Hostname: containerID,
		RootFS:   rootfs,
		Command:  config.Command,
		Args:     config.Args,
	})
	if err != nil {
		return fmt.Errorf("container error: %v", err)
	}

	// Step 3: Attach PID to cgroup
	err = container.AttachProcess(containerID, pid)
	if err != nil {
		return fmt.Errorf("attach error: %v", err)
	}

	// Step 4: Setup network — Joyal!
	err = network.SetupNetwork(pid, "eth0")
	if err != nil {
		return fmt.Errorf("network error: %v", err)
	}
	fmt.Printf("Container PID: %d\n", pid)

	// Step 5: Cleanup
	err = container.CleanupCgroup(containerID)
	if err != nil {
		return fmt.Errorf("cleanup error: %v", err)
	}

	fmt.Println("Container finished!")
	return nil
}