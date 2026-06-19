package main

import (
	"flag"
	"fmt"
	"mini-docker/cmd"
	"os"
)

func main() {

	// Define "run" subcommand
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)

	// Flags for run command
	memory := runCmd.Int("memory", 50, "memory limit in MB")
	cpu    := runCmd.Int("cpu", 30, "cpu limit in percent")
	detach := runCmd.Bool("d", false, "run in background")

	// Check if user typed a command
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./minidocker <command>")
		fmt.Println("Commands:")
		fmt.Println("  run [-memory MB] [-cpu PERCENT] [-d] <image> <command>")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "run":
		// Parse flags after "run"
		runCmd.Parse(os.Args[2:])

		// Get remaining args after flags
		args := runCmd.Args()
		if len(args) < 2 {
			fmt.Println("Usage: ./minidocker run <image> <command>")
			os.Exit(1)
		}

		image   := args[0] // alpine
		command := args[1] // /bin/sh
		cmdArgs := args[2:] // extra arguments

		// Build config
		config := cmd.RunConfig{
			Image:      image,
			Command:    command,
			Args:       cmdArgs,
			MemoryMB:   *memory,
			CPUPercent: *cpu,
			Detach:     *detach,
		}

		// Start container!
		err := cmd.Run(config)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Unknown command:", os.Args[1])
		fmt.Println("Available commands: run")
		os.Exit(1)
	}
}