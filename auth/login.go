package auth

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"patroncli/config"
	"patroncli/types"
	"strings"
)

// Handle `auth configure`
func Configure() {
	config.Configure()
}

// Handle `auth login`
func LoginCommand(args []string) {
	// Create a flag set for the `login` subcommand
	loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
	profileName := loginCmd.String("profile", "", "The profile name to login with")

	// Parse the flags
	loginCmd.Parse(args)

	// Ensure the profile flag is provided
	if *profileName == "" {
		fmt.Println("login requires --profile flag")
		loginCmd.Usage()
		os.Exit(1)
	}

	// Call the login logic with the provided profile name
	performLogin(*profileName)
}

// performLogin handles the actual login logic
func performLogin(profileName string) {
	// Prompt the user for a password and login
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	profilesPath := config.GetConfigPath()
	data, _ := os.ReadFile(profilesPath)
	var profiles []types.Profile
	_ = json.Unmarshal(data, &profiles)

	var profile types.Profile
	for _, p := range profiles {
		if p.Name == profileName {
			profile = p
			break
		}
	}

	if profile.Name == "" {
		fmt.Println("Profile not found")
		return
	}

	url := fmt.Sprintf("https://%s:%s/api/login", profile.IP, profile.Port)
	body := map[string]interface{}{
		"username": profile.Username,
		"password": password,
		"duration": profile.LoginTime,
	}
	bodyData, _ := json.Marshal(body)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer resp.Body.Close()

	respData, _ := io.ReadAll(resp.Body)
	var result map[string]string
	_ = json.Unmarshal(respData, &result)

	token := result["token"]
	if token == "" {
		fmt.Println("Failed to login: invalid response")
		return
	}

	// Save the token
	cred := types.Credential{
		Profile: profile.Name,
		IP:      profile.IP,
		Port:    profile.Port,
		Token:   token,
	}
	err = config.SaveCredential(cred)
	if err != nil {
		fmt.Println("Error saving credentials:", err)
	} else {
		fmt.Println("Login successful, token saved!")
	}
}
