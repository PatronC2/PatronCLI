package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"patroncli/types"
)

func TestGetConfigPath_UsesHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := GetConfigPath()
	want := filepath.Join(home, ".patron", "config")
	if got != want {
		t.Fatalf("unexpected config path: want %q got %q", want, got)
	}
}

func TestProfileExists(t *testing.T) {
	profiles := []types.Profile{{Name: "dev"}, {Name: "prod"}}
	if !profileExists("prod", profiles) {
		t.Fatal("expected profileExists to return true")
	}
	if profileExists("staging", profiles) {
		t.Fatal("expected profileExists to return false")
	}
}

func TestSaveProfile_CreatesAndAppends(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	p1 := types.Profile{Name: "dev", IP: "1.1.1.1", Port: "443", Username: "u1", LoginTime: 8}
	p2 := types.Profile{Name: "prod", IP: "2.2.2.2", Port: "8443", Username: "u2", LoginTime: 12}

	if err := SaveProfile(p1); err != nil {
		t.Fatalf("save profile 1 failed: %v", err)
	}
	if err := SaveProfile(p2); err != nil {
		t.Fatalf("save profile 2 failed: %v", err)
	}

	profiles := readProfilesFile(t, GetConfigPath())
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0].Name != "dev" || profiles[1].Name != "prod" {
		t.Fatalf("unexpected profiles order/content: %#v", profiles)
	}
}

func TestSaveProfile_UpdatesExisting(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	orig := types.Profile{Name: "dev", IP: "1.1.1.1", Port: "443", Username: "u1", LoginTime: 8}
	updated := types.Profile{
		Name:           "dev",
		IP:             "9.9.9.9",
		Port:           "9443",
		Username:       "new-user",
		LoginTime:      24,
		SOCKS5Enabled:  true,
		SOCKS5Host:     "127.0.0.1",
		SOCKS5Port:     "1080",
		SOCKS5Username: "proxy-user",
		SOCKS5Password: "proxy-pass",
	}

	if err := SaveProfile(orig); err != nil {
		t.Fatalf("save original profile failed: %v", err)
	}
	if err := SaveProfile(updated); err != nil {
		t.Fatalf("save updated profile failed: %v", err)
	}

	profiles := readProfilesFile(t, GetConfigPath())
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile after update, got %d", len(profiles))
	}
	got := profiles[0]
	if got.IP != "9.9.9.9" || got.Port != "9443" || got.Username != "new-user" || got.LoginTime != 24 {
		t.Fatalf("profile not updated correctly: %#v", got)
	}
	if !got.SOCKS5Enabled || got.SOCKS5Host != "127.0.0.1" || got.SOCKS5Port != "1080" {
		t.Fatalf("socks5 fields not updated correctly: %#v", got)
	}
}

func TestSaveProfile_InvalidExistingJSONIsOverwritten(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	configPath := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("not-json"), 0o644); err != nil {
		t.Fatalf("write invalid file failed: %v", err)
	}

	p := types.Profile{Name: "dev", IP: "1.1.1.1", Port: "443", Username: "u1", LoginTime: 8}
	if err := SaveProfile(p); err != nil {
		t.Fatalf("save after invalid json failed: %v", err)
	}

	profiles := readProfilesFile(t, configPath)
	if len(profiles) != 1 || profiles[0].Name != "dev" {
		t.Fatalf("expected overwritten config with one profile, got %#v", profiles)
	}
}

func readProfilesFile(t *testing.T, path string) []types.Profile {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config file failed: %v", err)
	}
	var profiles []types.Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		t.Fatalf("unmarshal profiles failed: %v", err)
	}
	return profiles
}
