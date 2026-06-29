package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"github.com/dheeraj20064/Mini-Docker/cmd"
)

func main() {
	// 1. Handle "child" mode: This is the re-exec entry point
	// If the first argument is "child", we are now running INSIDE the isolated namespaces.
	if len(os.Args) > 1 && os.Args[1] == "child" {
		if len(os.Args) < 4 {
			fmt.Println("Child mode requires: rootfs, hostname, and command")
			os.Exit(1)
		}
		
		rootfs := os.Args[2]
		hostname := os.Args[3]
		// The rest of the args are the command and its arguments
		command := os.Args[4:]

		// Setup the filesystem isolation
		err := setupRootFS(rootfs, hostname)
		if err != nil {
			fmt.Printf("Child setup error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("[Child] RootFS setup complete. Verifying execution environment...")
		
		containerPath := "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
		os.Setenv("PATH", containerPath)

		targetBinary := command[0]
		if !strings.Contains(targetBinary, "/") {
			// Resolve target binary against container PATH
			for _, dir := range strings.Split(containerPath, ":") {
				candidate := dir + "/" + targetBinary
				if _, err := os.Stat(candidate); err == nil {
					targetBinary = candidate
					break
				}
			}
		}

		if _, err := os.Stat(targetBinary); os.IsNotExist(err) {
			fmt.Printf("[Child] Warning: Binary %s not directly stat-able in rootfs.\n", targetBinary)
		} else {
			fmt.Printf("[Child] Found binary %s. Proceeding to execute...\n", targetBinary)
		}

		// Execute the final command
		cmd := exec.Command(targetBinary, command[1:]...)
		cmd.Env = []string{
			"PATH=" + containerPath,
			"TERM=xterm",
			"HOME=/root",
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Container process exited: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 2. Handle "parent" mode: The standard CLI entry point
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	memory := runCmd.Int("memory", 50, "memory limit in MB")
	cpu    := runCmd.Int("cpu", 30, "cpu limit in percent")
	detach := runCmd.Bool("d", false, "run in background")

	if len(os.Args) < 2 {
		fmt.Println("Usage: ./minidocker <command>")
		fmt.Println("Commands:")
		fmt.Println("  run [-memory MB] [-cpu PERCENT] [-d] <image> <command>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runCmd.Parse(os.Args[2:])
		args := runCmd.Args()
		if len(args) < 2 {
			fmt.Println("Usage: ./minidocker run <image> <command>")
			os.Exit(1)
		}

		image   := args[0]
		command := args[1]
		cmdArgs := args[2:]

		config := cmd.RunConfig{
			Image:      image,
			Command:    command,
			Args:       cmdArgs,
			MemoryMB:   *memory,
			CPUPercent: *cpu,
			Detach:     *detach,
		}

		err := cmd.Run(config)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

func setupRootFS(rootfs, hostname string) error {
	// Set hostname inside child namespace
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		return fmt.Errorf("sethostname: %v", err)
	}

	// Make root recursively private to decouple mount propagation and avoid EINVAL with pivot_root
	if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("mount private: %v", err)
	}

	// Bind mount rootfs to itself (pivot_root requirement)
	if err := syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("bind mount: %v", err)
	}

	// Ensure rootfs mount is also recursively private
	if err := syscall.Mount("", rootfs, "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("mount rootfs private: %v", err)
	}

	oldRoot := rootfs + "/old_root"
	if err := os.MkdirAll(oldRoot, 0700); err != nil {
		return fmt.Errorf("mkdir old_root: %v", err)
	}

	if err := syscall.PivotRoot(rootfs, oldRoot); err != nil {
		return fmt.Errorf("pivot_root: %v", err)
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("chdir /: %v", err)
	}

	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("mount proc: %v", err)
	}

	_ = syscall.Mount("sysfs", "/sys", "sysfs", 0, "")
	_ = syscall.Mount("tmpfs", "/dev", "tmpfs", 0, "")

	if err := syscall.Unmount("/old_root", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount old_root: %v", err)
	}
	os.RemoveAll("/old_root")

	return nil
}
