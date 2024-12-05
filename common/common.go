package common

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"patroncli/config"
	"patroncli/types"
)

// makeRequest is a generic function for API requests (GET, POST, PUT, DELETE)
func MakeRequest(method, url string, profile types.Profile, body interface{}) ([]byte, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Ignore self-signed certificates
			},
		},
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Authorization", GetProfileToken(profile.Name))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return responseData, nil
}

// GetProfileToken retrieves the token for a given profile
func GetProfileToken(profileName string) string {
	credentialsPath := config.GetCredentialsPath()
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		fmt.Printf("Warning: failed to read credentials file: %v\n", err)
		return ""
	}

	var credentials []types.Credential
	err = json.Unmarshal(data, &credentials)
	if err != nil {
		fmt.Printf("Warning: failed to parse credentials file: %v\n", err)
		return ""
	}

	for _, cred := range credentials {
		if cred.Profile == profileName {
			return cred.Token
		}
	}

	fmt.Printf("Warning: token not found for profile '%s'\n", profileName)
	return ""
}
