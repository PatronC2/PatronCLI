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
func MakeRequest(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
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
	req.Header.Set("Authorization", profile.Token)
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

func GetCreds(profileName string) types.Credential {
	credentialsPath := config.GetCredentialsPath()
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		fmt.Printf("Warning: failed to read credentials file: %v\n", err)
		return types.Credential{}
	}

	var creds []types.Credential
	err = json.Unmarshal(data, &creds)
	if err != nil {
		return types.Credential{}
	}

	for _, cred := range creds {
		if cred.Profile == profileName {
			return cred
		}
	}
	return types.Credential{}
}
