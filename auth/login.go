package auth

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"patroncli/common"
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

func performLogin(profileName string) {
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
	requestBody := map[string]interface{}{
		"username": profile.Username,
		"password": password,
		"duration": profile.LoginTime,
	}

	responseBody, err := common.MakeRequest("POST", url, profile, requestBody)
	if err != nil {
		fmt.Printf("Request error: %v\n", err)
		return
	}

	var result map[string]string
	if err := json.Unmarshal(responseBody, &result); err != nil {
		fmt.Println("Failed to parse login response")
		return
	}

	token := result["token"]
	if token == "" {
		fmt.Println("Failed to login: invalid response")
		return
	}

	cred := types.Credential{
		Profile: profile.Name,
		IP:      profile.IP,
		Port:    profile.Port,
		Token:   token,
	}
	if err := config.SaveCredential(cred); err != nil {
		fmt.Printf("Error saving credentials: %v\n", err)
	} else {
		fmt.Println("Login successful, token saved!")
	}
}
