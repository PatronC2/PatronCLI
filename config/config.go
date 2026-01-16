package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"patroncli/types"
	"strconv"
	"strings"
)

func GetConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ".patron", "config")
}

func SaveProfile(profile types.Profile) error {
	configPath := GetConfigPath()
	_ = os.MkdirAll(filepath.Dir(configPath), 0755)

	var profiles []types.Profile
	if _, err := os.Stat(configPath); err == nil {
		data, _ := os.ReadFile(configPath)
		json.Unmarshal(data, &profiles)
	}

	for i, p := range profiles {
		if p.Name == profile.Name {
			profiles[i] = profile
			break
		}
	}

	if !profileExists(profile.Name, profiles) {
		profiles = append(profiles, profile)
	}

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize profiles: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func profileExists(name string, profiles []types.Profile) bool {
	for _, p := range profiles {
		if p.Name == name {
			return true
		}
	}
	return false
}

func Configure() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter profile name: ")
	profileName, _ := reader.ReadString('\n')

	fmt.Print("Enter API IP: ")
	apiIP, _ := reader.ReadString('\n')

	fmt.Print("Enter API Port: ")
	apiPort, _ := reader.ReadString('\n')

	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter login time (default 8): ")
	loginTimeInput, _ := reader.ReadString('\n')
	loginTime, err := strconv.Atoi(strings.TrimSpace(loginTimeInput))
	if err != nil {
		loginTime = 8
	}

	// SOCKS5 proxy prompts
	fmt.Print("Use SOCKS5 proxy? (y/N): ")
	useProxyIn, _ := reader.ReadString('\n')
	useProxy := strings.EqualFold(strings.TrimSpace(useProxyIn), "y") ||
		strings.EqualFold(strings.TrimSpace(useProxyIn), "yes")

	var proxyHost, proxyPort, proxyUser, proxyPass string
	if useProxy {
		fmt.Print("SOCKS5 proxy host (e.g. 127.0.0.1): ")
		proxyHost, _ = reader.ReadString('\n')

		fmt.Print("SOCKS5 proxy port (e.g. 1080): ")
		proxyPort, _ = reader.ReadString('\n')

		fmt.Print("SOCKS5 username (blank for none): ")
		proxyUser, _ = reader.ReadString('\n')

		if strings.TrimSpace(proxyUser) != "" {
			fmt.Print("SOCKS5 password: ")
			proxyPass, _ = reader.ReadString('\n')
		}
	}

	profile := types.Profile{
		Name:      strings.TrimSpace(profileName),
		IP:        strings.TrimSpace(apiIP),
		Port:      strings.TrimSpace(apiPort),
		Username:  strings.TrimSpace(username),
		LoginTime: loginTime,

		SOCKS5Enabled:  useProxy,
		SOCKS5Host:     strings.TrimSpace(proxyHost),
		SOCKS5Port:     strings.TrimSpace(proxyPort),
		SOCKS5Username: strings.TrimSpace(proxyUser),
		SOCKS5Password: strings.TrimSpace(proxyPass),
	}

	err = SaveProfile(profile)
	if err != nil {
		fmt.Println("Error saving profile:", err)
	} else {
		fmt.Println("Profile saved successfully!")
	}
}
