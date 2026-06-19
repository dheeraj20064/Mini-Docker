package container

import (
	"fmt"
	"os"
	"strconv"
)

// CgroupConfig holds resource limit settings
type CgroupConfig struct {
	ContainerID string
	MemoryMB    int
	CPUPercent  int
}

// SetupCgroup creates cgroup folder and sets limits
func SetupCgroup(config CgroupConfig) error {

	cgroupPath := "/sys/fs/cgroup/minidocker-" + config.ContainerID

	// Create cgroup folder
	err := os.MkdirAll(cgroupPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create cgroup: %v", err)
	}

	// Set memory limit
	memoryBytes := config.MemoryMB * 1024 * 1024
	err = os.WriteFile(
		cgroupPath+"/memory.max",
		[]byte(strconv.Itoa(memoryBytes)),
		0700,
	)
	if err != nil {
		return fmt.Errorf("failed to set memory limit: %v", err)
	}

	// Set CPU limit
	cpuMax := fmt.Sprintf("%d 100000", config.CPUPercent*1000)
	err = os.WriteFile(
		cgroupPath+"/cpu.max",
		[]byte(cpuMax),
		0700,
	)
	if err != nil {
		return fmt.Errorf("failed to set cpu limit: %v", err)
	}

	fmt.Printf("Cgroup ready: memory=%dMB cpu=%d%%\n",
		config.MemoryMB, config.CPUPercent)
	return nil
}

// AttachProcess attaches container PID to cgroup
func AttachProcess(containerID string, pid int) error {
	cgroupPath := "/sys/fs/cgroup/minidocker-" + containerID

	err := os.WriteFile(
		cgroupPath+"/cgroup.procs",
		[]byte(strconv.Itoa(pid)),
		0700,
	)
	if err != nil {
		return fmt.Errorf("failed to attach process: %v", err)
	}

	fmt.Printf("Process %d attached to cgroup\n", pid)
	return nil
}

// CleanupCgroup removes cgroup after container stops
func CleanupCgroup(containerID string) error {
	cgroupPath := "/sys/fs/cgroup/minidocker-" + containerID

	err := os.RemoveAll(cgroupPath)
	if err != nil {
		return fmt.Errorf("failed to cleanup cgroup: %v", err)
	}

	fmt.Println("Cgroup cleaned up!")
	return nil
}