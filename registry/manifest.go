package registry

import (
	"encoding/json"
	"fmt"
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

	req.Header.Set(
		"Accept",
		"application/vnd.docker.distribution.manifest.v2+json",
	)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest request failed: %s", resp.Status)
	}

	var manifest Manifest

	err = json.NewDecoder(resp.Body).Decode(&manifest)

	if err != nil {
		return nil, err
	}

	return &manifest, nil
}
