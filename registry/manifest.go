package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetManifest(repo string, tag string, token string) (*Manifest, error) {

	url := fmt.Sprintf(
		"https://registry-1.docker.io/v2/%s/manifests/%s",
		repo, tag,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Accept BOTH single and multi-arch manifest formats
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	req.Header.Add("Accept", "application/vnd.oci.image.manifest.v1+json")
	req.Header.Add("Accept", "application/vnd.oci.image.index.v1+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest request failed: %s", resp.Status)
	}

	// Read body once so we can try decoding it two ways
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Try manifest list first (multi-arch)
	var list ManifestList
	_ = json.Unmarshal(body, &list)

	if list.MediaType == "application/vnd.docker.distribution.manifest.list.v2+json" ||
    list.MediaType == "application/vnd.oci.image.index.v1+json" {
		// Find linux/amd64 and fetch its specific manifest
		for _, m := range list.Manifests {
			if m.Platform.OS == "linux" && m.Platform.Architecture == "amd64" {
				fmt.Printf("[registry] multi-arch image, selecting linux/amd64: %s\n", m.Digest)
				return GetManifest(repo, m.Digest, token)
			}
		}
		return nil, fmt.Errorf("no linux/amd64 manifest found")
	}

	// Single-arch manifest
	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, err
	}

	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("manifest has zero layers")
	}

	return &manifest, nil
}
