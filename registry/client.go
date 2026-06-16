package registry

import (
	"fmt"
	"os"
)

func cacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/minidocker-cache"
	}
	return home + "/.minidocker/cache"
}

func PullImage(image string) ([]string, error) {
	repo, tag := ParseImage(image)
	fmt.Printf("[registry] pulling %s (tag: %s)\n", repo, tag)

	token, err := GetToken(repo)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	manifest, err := GetManifest(repo, tag, token)
	if err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}

	fmt.Printf("[registry] found %d layer(s)\n", len(manifest.Layers))

	dir := cacheDir()
	paths := make([]string, 0, len(manifest.Layers))

	for i, layer := range manifest.Layers {
		fmt.Printf("[registry] layer %d/%d: %s\n", i+1, len(manifest.Layers), layer.Digest[:19])
		p, err := DownloadLayer(repo, layer.Digest, token, dir)
		if err != nil {
			return nil, fmt.Errorf("download layer %s: %w", layer.Digest, err)
		}
		paths = append(paths, p)
	}

	return paths, nil
}