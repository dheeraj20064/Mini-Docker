package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	BaseDir    = "/var/lib/minidocker"
	LayersDir  = BaseDir + "/layers"
	MergedDir  = BaseDir + "/merged"
	UpperDir   = BaseDir + "/upper"
	WorkDir    = BaseDir + "/work"
)

// InitializeStorage ensures the base directory structure exists
func InitializeStorage() error {
	dirs := []string{LayersDir, MergedDir, UpperDir, WorkDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// PrepareRootFS takes extracted layer paths and sets up an OverlayFS mount for a container
func PrepareRootFS(containerID string, layerPaths []string) (string, error) {
	if err := InitializeStorage(); err != nil {
		return "", err
	}

	// 1. Extract all provided layers into the LayersDir
	var lowerDirs []string
	for i, path := range layerPaths {
		layerID := strings.TrimSuffix(filepath.Base(path), ".tar.gz")
		if layerID == "" {
			layerID = fmt.Sprintf("layer-%d", i)
		}
		dest := filepath.Join(LayersDir, layerID)

		// Skip extraction if already exists
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			if err := ExtractLayer(path, dest); err != nil {
				return "", fmt.Errorf("extraction failed for layer %d: %w", i, err)
			}
		}
		lowerDirs = append(lowerDirs, dest)
	}

	// 2. Setup container-specific directories
	containerMerged := filepath.Join(MergedDir, containerID)
	containerUpper := filepath.Join(UpperDir, containerID)
	containerWork := filepath.Join(WorkDir, containerID)

	if err := os.MkdirAll(containerMerged, 0755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(containerUpper, 0755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(containerWork, 0755); err != nil {
		return "", err
	}

	// 3. Mount using OverlayFS
	// lowerdir is a colon-separated list of directories
	lowerDirsJoined := strings.Join(lowerDirs, ":")
	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDirsJoined, containerUpper, containerWork)

	if err := MountOverlay(containerMerged, options); err != nil {
		return "", fmt.Errorf("overlay mount failed: %w", err)
	}

	return containerMerged, nil
}

// CleanupRootFS unmounts the merged directory and removes container-specific data
func CleanupRootFS(containerID string) error {
	merged := filepath.Join(MergedDir, containerID)
	
	// Unmount the merged directory
	// Using MNT_DETACH (lazy unmount) to ensure it's cleaned up even if busy
	// We use a shell command or syscall.Unmount. Since we used syscall.Mount, we'll use syscall.Unmount if possible, 
	// but often 'umount -l' is more reliable in these scripts.
	
	// For simplicity and robustness in this implementation, we'll use the syscall.
	// However, we'll define it in overlay.go
	if err := Unmount(merged); err != nil {
		return fmt.Errorf("unmount failed: %w", err)
	}

	// Remove the upper and merged directories
	os.RemoveAll(filepath.Join(UpperDir, containerID))
	os.RemoveAll(filepath.Join(WorkDir, containerID))
	os.RemoveAll(merged)

	return nil
}
