package registry

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func DownloadLayer(repo, digest, token, cacheDir string) (string, error) {

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}

	filename := strings.ReplaceAll(digest, ":", "_") + ".tar.gz"
	path := filepath.Join(cacheDir, filename)

	// Cache hit — skip download
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("[registry] cached: %s\n", filepath.Base(path))
		return path, nil
	}

	url := fmt.Sprintf(
		"https://registry-1.docker.io/v2/%s/blobs/%s",
		repo, digest,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("blob fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("blob request failed: %s", resp.Status)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return "", fmt.Errorf("write blob: %w", err)
	}

	fmt.Printf("[registry] downloaded: %s (%d bytes)\n", filepath.Base(path), written)
	return path, nil
}