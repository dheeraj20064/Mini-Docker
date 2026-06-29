package storage

import (
	"fmt"
	"syscall"
)

// MountOverlay executes the mount system call for OverlayFS
func MountOverlay(target, options string) error {
	// syscall.Mount(source, target, filesystemtype, mountflags, data)
	// For OverlayFS, the 'source' is actually ignored/empty, and 
	// the filesystem configuration is passed via the 'data' parameter.
	err := syscall.Mount("overlay", target, "overlay", 0, options)
	if err != nil {
		return fmt.Errorf("mount syscall failed: %w", err)
	}
	return nil
}

// Unmount removes the mount point from the system
func Unmount(target string) error {
	// MNT_DETACH is used for a "lazy" unmount, which helps when 
	// the mount point is still being accessed by the container.
	err := syscall.Unmount(target, syscall.MNT_DETACH)
	if err != nil {
		return fmt.Errorf("unmount syscall failed: %w", err)
	}
	return nil
}
