package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"patroncli/types"
)

func TestGetCredentialsPath_UsesHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := GetCredentialsPath()
	want := filepath.Join(home, ".patron", "credentials")
	if got != want {
		t.Fatalf("unexpected credentials path: want %q got %q", want, got)
	}
}

func TestSaveCredential_CreatesAndAppends(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	c1 := types.Credential{Profile: "dev", Token: "token-1", IP: "1.1.1.1", Port: "443"}
	c2 := types.Credential{Profile: "prod", Token: "token-2", IP: "2.2.2.2", Port: "8443"}

	if err := SaveCredential(c1); err != nil {
		t.Fatalf("save credential 1 failed: %v", err)
	}
	if err := SaveCredential(c2); err != nil {
		t.Fatalf("save credential 2 failed: %v", err)
	}

	creds := readCredentialsFile(t, GetCredentialsPath())
	if len(creds) != 2 {
		t.Fatalf("expected 2 credentials, got %d", len(creds))
	}
	if creds[0].Profile != "dev" || creds[1].Profile != "prod" {
		t.Fatalf("unexpected credentials content: %#v", creds)
	}
}

func TestSaveCredential_UpdatesExisting(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	orig := types.Credential{Profile: "dev", Token: "old", IP: "1.1.1.1", Port: "443"}
	updated := types.Credential{
		Profile:        "dev",
		Token:          "new-token",
		IP:             "9.9.9.9",
		Port:           "9443",
		SOCKS5Enabled:  true,
		SOCKS5Host:     "127.0.0.1",
		SOCKS5Port:     "1080",
		SOCKS5Username: "proxy-user",
		SOCKS5Password: "proxy-pass",
	}

	if err := SaveCredential(orig); err != nil {
		t.Fatalf("save original credential failed: %v", err)
	}
	if err := SaveCredential(updated); err != nil {
		t.Fatalf("save updated credential failed: %v", err)
	}

	creds := readCredentialsFile(t, GetCredentialsPath())
	if len(creds) != 1 {
		t.Fatalf("expected 1 credential after update, got %d", len(creds))
	}
	got := creds[0]
	if got.Token != "new-token" || got.IP != "9.9.9.9" || got.Port != "9443" {
		t.Fatalf("credential not updated correctly: %#v", got)
	}
	if !got.SOCKS5Enabled || got.SOCKS5Host != "127.0.0.1" || got.SOCKS5Port != "1080" {
		t.Fatalf("socks5 fields not updated correctly: %#v", got)
	}
}

func TestSaveCredential_InvalidExistingJSONReturnsError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	credPath := GetCredentialsPath()
	if err := os.MkdirAll(filepath.Dir(credPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(credPath, []byte("not-json"), 0o644); err != nil {
		t.Fatalf("write invalid json failed: %v", err)
	}

	err := SaveCredential(types.Credential{Profile: "dev", Token: "x"})
	if err == nil {
		t.Fatal("expected error for invalid credentials file")
	}
	if !strings.Contains(err.Error(), "failed to parse credentials file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSaveCredential_ExistingPathIsDirectoryReturnsReadError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	credPath := GetCredentialsPath()
	if err := os.MkdirAll(credPath, 0o755); err != nil {
		t.Fatalf("mkdir cred path failed: %v", err)
	}

	err := SaveCredential(types.Credential{Profile: "dev", Token: "x"})
	if err == nil {
		t.Fatal("expected error when credentials path is directory")
	}
	if !strings.Contains(err.Error(), "failed to read credentials file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func readCredentialsFile(t *testing.T, path string) []types.Credential {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read credentials file failed: %v", err)
	}
	var creds []types.Credential
	if err := json.Unmarshal(data, &creds); err != nil {
		t.Fatalf("unmarshal credentials failed: %v", err)
	}
	return creds
}
