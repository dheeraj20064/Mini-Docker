package storage

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ExtractLayer decompresses a .tar.gz blob into the destination directory
func ExtractLayer(src, dest string) error {
	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("mkdir destination: %w", err)
	}

	// Open the .tar.gz file
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source tarball: %w", err)
	}
	defer f.Close()

	// Wrap in a gzip reader
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gzr.Close()

	// Wrap in a tar reader
	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}

		// Sanitize the target path to prevent ZipSlip (Directory Traversal)
		target := filepath.Join(dest, header.Name)
		if !filepath.HasPrefix(target, filepath.Clean(dest)) {
			return fmt.Errorf("illegal file path in tarball: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
		case tar.TypeReg:
			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("mkdir parent: %w", err)
			}
			// Create file
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("copy file content: %w", err)
			}
			f.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("mkdir parent for symlink: %w", err)
			}
			os.Remove(target)
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("symlink (%s -> %s): %w", target, header.Linkname, err)
			}
		case tar.TypeLink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("mkdir parent for hardlink: %w", err)
			}
			oldPath := filepath.Join(dest, header.Linkname)
			os.Remove(target)
			if err := os.Link(oldPath, target); err != nil {
				return fmt.Errorf("hardlink (%s -> %s): %w", target, oldPath, err)
			}
		default:
			// Skip unknown flags
		}
	}

	return nil
}
