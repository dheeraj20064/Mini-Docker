package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// ContainerConfig holds container settings
type ContainerConfig struct {
	ID       string
	Hostname string
	RootFS   string // comes from Ameen (Student 2)!
	Command  string
	Args     []string
}

// StartContainer starts an isolated container process
func StartContainer(config ContainerConfig) (int, error) {

	// Create the command to run inside container
	cmd := exec.Command(config.Command, config.Args...)

	// Connect to terminal
	cmd.Stdin  = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Add isolation walls using syscall!
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | // own hostname
					syscall.CLONE_NEWPID | // own processes
					syscall.CLONE_NEWNS,   // own filesystem
	}

	// Start the container
	err := cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("failed to start container: %v", err)
	}

	// Set container hostname
	err = syscall.Sethostname([]byte(config.Hostname))
	if err != nil {
		return 0, fmt.Errorf("failed to set hostname: %v", err)
	}

	// Get PID — send this to Joyal (Student 3)!
	pid := cmd.Process.Pid
	fmt.Printf("Container started! PID: %d\n", pid)

	// Setup filesystem if RootFS provided by Ameen
	if config.RootFS != "" {
		err = setupRootFS(config.RootFS)
		if err != nil {
			return 0, fmt.Errorf("failed to setup rootfs: %v", err)
		}
	}

	// Wait for container to finish
	err = cmd.Wait()
	if err != nil {
		return pid, fmt.Errorf("container exited with error: %v", err)
	}

	return pid, nil
}

// setupRootFS switches container to Alpine filesystem
func setupRootFS(rootfs string) error {

	// Mount rootfs on itself first
	err := syscall.Mount(
		rootfs,
		rootfs,
		"",
		syscall.MS_BIND|syscall.MS_REC,
		"",
	)
	if err != nil {
		return fmt.Errorf("failed to bind mount: %v", err)
	}

	// Create old_root folder inside rootfs
	oldRoot := rootfs + "/old_root"
	err = os.MkdirAll(oldRoot, 0700)
	if err != nil {
		return fmt.Errorf("failed to create old_root: %v", err)
	}

	// Switch root filesystem!
	err = syscall.PivotRoot(rootfs, oldRoot)
	if err != nil {
		return fmt.Errorf("failed to pivot_root: %v", err)
	}

	// Go to new root
	err = os.Chdir("/")
	if err != nil {
		return fmt.Errorf("failed to chdir: %v", err)
	}

	// Mount /proc so ps aux works!
	err = syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		return fmt.Errorf("failed to mount proc: %v", err)
	}

	// Remove old root — not needed anymore!
	err = syscall.Unmount("/old_root", syscall.MNT_DETACH)
	if err != nil {
		return fmt.Errorf("failed to unmount old root: %v", err)
	}

	err = os.RemoveAll("/old_root")
	if err != nil {
		return fmt.Errorf("failed to remove old root: %v", err)
	}

	fmt.Println("Filesystem switched to Alpine!")
	return nil
}