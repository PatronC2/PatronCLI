package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"patroncli/types"
)

func GetCredentialsPath() string {
	return filepath.Join(os.Getenv("HOME"), ".patron", "credentials")
}

func SaveCredential(newCred types.Credential) error {
	credPath := GetCredentialsPath()
	_ = os.MkdirAll(filepath.Dir(credPath), 0755)

	var credentials []types.Credential

	if _, err := os.Stat(credPath); err == nil {
		data, err := os.ReadFile(credPath)
		if err != nil {
			return fmt.Errorf("failed to read credentials file: %w", err)
		}
		if err := json.Unmarshal(data, &credentials); err != nil {
			return fmt.Errorf("failed to parse credentials file: %w", err)
		}
	}

	updated := false
	for i, cred := range credentials {
		if cred.Profile == newCred.Profile {
			credentials[i] = newCred
			updated = true
			break
		}
	}

	if !updated {
		credentials = append(credentials, newCred)
	}

	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize credentials: %w", err)
	}

	return os.WriteFile(credPath, data, 0644)
}
