package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type tokenResponse struct {
	Token string `json:"token"`
}

// GetToken fetches a Bearer token for the given repository from Docker Hub.
func GetToken(repository string) (string, error) {
	url := fmt.Sprintf(
		"https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull",
		repository,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("token fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed: %s", resp.Status)
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("token decode: %w", err)
	}
	if tr.Token == "" {
		return "", fmt.Errorf("empty token received")
	}
	return tr.Token, nil
}
