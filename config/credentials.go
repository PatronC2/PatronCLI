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

func SaveCredential(cred types.Credential) error {
	credPath := GetCredentialsPath()
	_ = os.MkdirAll(filepath.Dir(credPath), 0755)

	var credentials []types.Credential
	if _, err := os.Stat(credPath); err == nil {
		data, _ := os.ReadFile(credPath)
		json.Unmarshal(data, &credentials)
	}

	credentials = append(credentials, cred)

	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize credentials: %w", err)
	}

	return os.WriteFile(credPath, data, 0644)
}
