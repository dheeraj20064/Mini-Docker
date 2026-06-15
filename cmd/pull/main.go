package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dheeraj20064/Mini-Docker/registry"
)

func main() {
	image := "alpine:latest"
	if len(os.Args) > 1 {
		image = os.Args[1]
	}

	paths, err := registry.PullImage(image)
	if err != nil {
		log.Fatalf("pull failed: %v", err)
	}

	fmt.Println("\n[done] layer paths for Student 2:")
	for _, p := range paths {
		fmt.Println(" ", p)
	}
}