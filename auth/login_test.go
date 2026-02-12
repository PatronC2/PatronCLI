package auth

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"patroncli/types"
)

func sampleProfile() types.Profile {
	return types.Profile{
		Name:           "p1",
		IP:             "127.0.0.1",
		Port:           "8443",
		Username:       "admin",
		LoginTime:      8,
		SOCKS5Enabled:  true,
		SOCKS5Host:     "10.0.0.1",
		SOCKS5Port:     "1080",
		SOCKS5Username: "proxyu",
		SOCKS5Password: "proxyp",
	}
}

func TestRunLogin_ProfileNotFound(t *testing.T) {
	deps := LoginDeps{
		ReadProfiles: func() ([]types.Profile, error) { return []types.Profile{}, nil },
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return nil, nil
		},
		SaveCredential: func(types.Credential) error { return nil },
	}

	err := runLogin("missing", "secret", deps)
	if err == nil || err.Error() != "profile not found" {
		t.Fatalf("expected profile not found, got %v", err)
	}
}

func TestRunLogin_RequestError(t *testing.T) {
	deps := LoginDeps{
		ReadProfiles: func() ([]types.Profile, error) { return []types.Profile{sampleProfile()}, nil },
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return nil, errors.New("boom")
		},
		SaveCredential: func(types.Credential) error { return nil },
	}

	err := runLogin("p1", "secret", deps)
	if err == nil || !strings.Contains(err.Error(), "request error") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestRunLogin_InvalidJSON(t *testing.T) {
	deps := LoginDeps{
		ReadProfiles: func() ([]types.Profile, error) { return []types.Profile{sampleProfile()}, nil },
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte("not-json"), nil
		},
		SaveCredential: func(types.Credential) error { return nil },
	}

	err := runLogin("p1", "secret", deps)
	if err == nil || err.Error() != "failed to parse login response" {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestRunLogin_MissingToken(t *testing.T) {
	deps := LoginDeps{
		ReadProfiles: func() ([]types.Profile, error) { return []types.Profile{sampleProfile()}, nil },
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"ok":"true"}`), nil
		},
		SaveCredential: func(types.Credential) error { return nil },
	}

	err := runLogin("p1", "secret", deps)
	if err == nil || err.Error() != "failed to login: invalid response" {
		t.Fatalf("expected missing token error, got %v", err)
	}
}

func TestRunLogin_SaveCredentialError(t *testing.T) {
	deps := LoginDeps{
		ReadProfiles: func() ([]types.Profile, error) { return []types.Profile{sampleProfile()}, nil },
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"token":"abc"}`), nil
		},
		SaveCredential: func(types.Credential) error { return errors.New("disk full") },
	}

	err := runLogin("p1", "secret", deps)
	if err == nil || !strings.Contains(err.Error(), "error saving credentials") {
		t.Fatalf("expected save credential error, got %v", err)
	}
}

func TestRunLogin_Success(t *testing.T) {
	var gotMethod string
	var gotURL string
	var gotRequestBody map[string]interface{}
	var saved types.Credential

	deps := LoginDeps{
		ReadProfiles: func() ([]types.Profile, error) { return []types.Profile{sampleProfile()}, nil },
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotMethod = method
			gotURL = url
			cast, ok := body.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("unexpected body type %T", body)
			}
			gotRequestBody = cast
			if gotRequestBody["password"] != "secret" {
				return nil, fmt.Errorf("password mismatch")
			}
			return []byte(`{"token":"abc123"}`), nil
		},
		SaveCredential: func(c types.Credential) error {
			saved = c
			return nil
		},
	}

	err := runLogin("p1", "secret", deps)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotMethod != "POST" {
		t.Fatalf("expected method POST, got %q", gotMethod)
	}
	if gotURL != "https://127.0.0.1:8443/api/login" {
		t.Fatalf("unexpected url: %q", gotURL)
	}
	if saved.Profile != "p1" || saved.Token != "abc123" {
		t.Fatalf("unexpected saved credential: %#v", saved)
	}
	if !saved.SOCKS5Enabled || saved.SOCKS5Host != "10.0.0.1" {
		t.Fatalf("expected SOCKS5 fields copied, got %#v", saved)
	}
}

func TestRunLoginCommand_MissingProfile(t *testing.T) {
	deps := LoginDeps{}
	var out bytes.Buffer
	err := runLoginCommand([]string{}, deps, strings.NewReader("ignored\n"), &out)
	if !errors.Is(err, errLoginUsage) {
		t.Fatalf("expected errLoginUsage, got %v", err)
	}
	if !strings.Contains(out.String(), "login requires --profile flag") {
		t.Fatalf("expected usage output, got %q", out.String())
	}
}

func TestRunLoginCommand_Success(t *testing.T) {
	deps := LoginDeps{
		ReadProfiles: func() ([]types.Profile, error) { return []types.Profile{sampleProfile()}, nil },
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"token":"abc"}`), nil
		},
		SaveCredential: func(types.Credential) error { return nil },
	}
	var out bytes.Buffer

	err := runLoginCommand([]string{"--profile", "p1"}, deps, strings.NewReader("secret\n"), &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !strings.Contains(out.String(), "Enter password:") || !strings.Contains(out.String(), "Login successful") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}
