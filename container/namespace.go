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

// StartContainer starts an isolated container process using the re-exec pattern
func StartContainer(config ContainerConfig) (*exec.Cmd, int, error) {
	// Get the path to the current executable
	exe, err := os.Executable()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get executable path: %v", err)
	}

	// We execute our own binary again, but with the "child" argument.
	// This ensures the setup (pivot_root, sethostname) happens INSIDE the new namespaces.
	args := []string{"child", config.RootFS, config.Hostname}
	args = append(args, config.Command)
	args = append(args, config.Args...)

	cmd := exec.Command(exe, args...)

	// Connect to terminal
	cmd.Stdin  = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Add isolation walls
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
					syscall.CLONE_NEWPID |
					syscall.CLONE_NEWNS |
					syscall.CLONE_NEWNET, // Added network namespace for Joyal's layer
	}

	// Start the container
	err = cmd.Start()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to start container child: %v", err)
	}

	pid := cmd.Process.Pid
	fmt.Printf("Container started! PID: %d\n", pid)

	return cmd, pid, nil
}

// WaitContainer waits for the container process to complete
func WaitContainer(cmd *exec.Cmd) error {
	if cmd == nil {
		return nil
	}
	return cmd.Wait()
}