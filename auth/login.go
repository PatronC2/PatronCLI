package auth

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"patroncli/common"
	"patroncli/config"
	"patroncli/types"
	"strings"
)

var errLoginUsage = errors.New("login usage error")

type LoginDeps struct {
	ReadProfiles   func() ([]types.Profile, error)
	MakeRequest    func(method, url string, profile types.Credential, body interface{}) ([]byte, error)
	SaveCredential func(types.Credential) error
}

func defaultLoginDeps() LoginDeps {
	return LoginDeps{
		ReadProfiles: func() ([]types.Profile, error) {
			profilesPath := config.GetConfigPath()
			data, err := os.ReadFile(profilesPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read profiles: %w", err)
			}
			var profiles []types.Profile
			if err := json.Unmarshal(data, &profiles); err != nil {
				return nil, fmt.Errorf("failed to parse profiles: %w", err)
			}
			return profiles, nil
		},
		MakeRequest: common.MakeRequest,
		SaveCredential: func(c types.Credential) error {
			return config.SaveCredential(c)
		},
	}
}

// Handle `auth configure`.
func Configure() {
	config.Configure()
}

// Handle `auth login`.
func LoginCommand(args []string) {
	if err := runLoginCommand(args, defaultLoginDeps(), os.Stdin, os.Stdout); err != nil {
		if errors.Is(err, errLoginUsage) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, err)
	}
}

func runLoginCommand(args []string, deps LoginDeps, in io.Reader, out io.Writer) error {
	loginCmd := flag.NewFlagSet("login", flag.ContinueOnError)
	loginCmd.SetOutput(out)
	profileName := loginCmd.String("profile", "", "The profile name to login with")

	if err := loginCmd.Parse(args); err != nil {
		return fmt.Errorf("%w: %v", errLoginUsage, err)
	}

	if *profileName == "" {
		fmt.Fprintln(out, "login requires --profile flag")
		loginCmd.Usage()
		return errLoginUsage
	}

	password, err := readPassword(in, out)
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	if err := runLogin(*profileName, password, deps); err != nil {
		return err
	}

	fmt.Fprintln(out, "Login successful, token saved!")
	return nil
}

func readPassword(in io.Reader, out io.Writer) (string, error) {
	reader := bufio.NewReader(in)
	fmt.Fprint(out, "Enter password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}

func runLogin(profileName, password string, deps LoginDeps) error {
	profiles, err := deps.ReadProfiles()
	if err != nil {
		return err
	}

	var profile types.Profile
	for _, p := range profiles {
		if p.Name == profileName {
			profile = p
			break
		}
	}

	if profile.Name == "" {
		return errors.New("profile not found")
	}

	url := fmt.Sprintf("https://%s:%s/api/login", profile.IP, profile.Port)
	requestBody := map[string]interface{}{
		"username": profile.Username,
		"password": password,
		"duration": profile.LoginTime,
	}

	headers := types.Credential{
		Profile:        profile.Name,
		IP:             profile.IP,
		Port:           profile.Port,
		SOCKS5Enabled:  profile.SOCKS5Enabled,
		SOCKS5Host:     profile.SOCKS5Host,
		SOCKS5Port:     profile.SOCKS5Port,
		SOCKS5Username: profile.SOCKS5Username,
		SOCKS5Password: profile.SOCKS5Password,
	}

	responseBody, err := deps.MakeRequest("POST", url, headers, requestBody)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	var result map[string]string
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return errors.New("failed to parse login response")
	}

	token := result["token"]
	if token == "" {
		return errors.New("failed to login: invalid response")
	}

	cred := types.Credential{
		Profile:        profile.Name,
		IP:             profile.IP,
		Port:           profile.Port,
		Token:          token,
		SOCKS5Enabled:  profile.SOCKS5Enabled,
		SOCKS5Host:     profile.SOCKS5Host,
		SOCKS5Port:     profile.SOCKS5Port,
		SOCKS5Username: profile.SOCKS5Username,
		SOCKS5Password: profile.SOCKS5Password,
	}
	if err := deps.SaveCredential(cred); err != nil {
		return fmt.Errorf("error saving credentials: %w", err)
	}

	return nil
}
