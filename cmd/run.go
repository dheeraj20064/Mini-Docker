package cmd

import (
	"crypto/rand"
	"fmt"
	"github.com/dheeraj20064/Mini-Docker/container"
	"github.com/dheeraj20064/Mini-Docker/network"
	"github.com/dheeraj20064/Mini-Docker/registry"
	"github.com/dheeraj20064/Mini-Docker/storage"
)

// RunConfig defines the configuration for running a container
type RunConfig struct {
	Image      string
	Command    string
	Args       []string
	MemoryMB   int
	CPUPercent int
	Detach     bool
}

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

	// Step 0: Pull image and prepare rootfs
	fmt.Println("[Pipeline] Pulling image and preparing filesystem...")
	layerPaths, err := registry.PullImage(config.Image)
	if err != nil {
		return fmt.Errorf("registry pull failed: %v", err)
	}

	rootfs, err := storage.PrepareRootFS(containerID, layerPaths)
	if err != nil {
		return fmt.Errorf("storage preparation failed: %v", err)
	}
	// Ensure filesystem is cleaned up after the container exits
	defer storage.CleanupRootFS(containerID)

	// Step 1: Setup cgroup
	err = container.SetupCgroup(container.CgroupConfig{
		ContainerID: containerID,
		MemoryMB:    config.MemoryMB,
		CPUPercent:  config.CPUPercent,
	})
	if err != nil {
		return fmt.Errorf("cgroup error: %v", err)
	}
	defer container.CleanupCgroup(containerID)

	// Step 2: Start isolated container child process
	cmdProc, pid, err := container.StartContainer(container.ContainerConfig{
		ID:       containerID,
		Hostname: containerID,
		RootFS:   rootfs,
		Command:  config.Command,
		Args:     config.Args,
	})
	if err != nil {
		return fmt.Errorf("container error: %v", err)
	}

	// Ensure network cleanup on exit
	defer network.CleanupContainerNetwork(pid, "eth0")

	// Step 3: Attach PID to cgroup while process is running
	err = container.AttachProcess(containerID, pid)
	if err != nil {
		fmt.Printf("[Warning] Failed to attach process to cgroup: %v\n", err)
	}

	// Step 4: Setup network for active container PID
	err = network.SetupNetwork(pid, "eth0")
	if err != nil {
		fmt.Printf("[Warning] Failed to setup network: %v\n", err)
	}
	fmt.Printf("Container PID: %d\n", pid)

	// Step 5: Wait for container to exit
	if err := container.WaitContainer(cmdProc); err != nil {
		fmt.Printf("Container process execution completed with error: %v\n", err)
	} else {
		fmt.Println("Container finished successfully!")
	}

	return nil
}
