package admin

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"patroncli/types"
)

func TestRunUpdateUserCommand_PasswordOnly(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotMethod, gotURL string
	var gotBody map[string]interface{}

	deps := updateUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "172.16.200.185", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotMethod = method
			gotURL = url
			gotBody = body.(map[string]interface{})
			return []byte(`{"message":"User updated successfully"}`), nil
		},
	}

	var out bytes.Buffer
	err := runUpdateUserCommand([]string{"--profile", "prod", "--username", "test", "--new-password", "newpassword"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotMethod != "PUT" {
		t.Fatalf("expected PUT method, got %q", gotMethod)
	}
	if gotURL != "https://172.16.200.185:8443/api/admin/users/test" {
		t.Fatalf("unexpected URL: %q", gotURL)
	}
	want := map[string]interface{}{"newPassword": "newpassword"}
	if !reflect.DeepEqual(gotBody, want) {
		t.Fatalf("unexpected body\nwant: %#v\ngot:  %#v", want, gotBody)
	}
}

func TestRunUpdateUserCommand_RoleOnly(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotBody map[string]interface{}
	deps := updateUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "172.16.200.185", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotBody = body.(map[string]interface{})
			return []byte(`{"message":"User updated successfully"}`), nil
		},
	}

	var out bytes.Buffer
	err := runUpdateUserCommand([]string{"--profile", "prod", "--username", "test", "--new-role", "operator"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	want := map[string]interface{}{"newRole": "operator"}
	if !reflect.DeepEqual(gotBody, want) {
		t.Fatalf("unexpected body\nwant: %#v\ngot:  %#v", want, gotBody)
	}
}

func TestRunUpdateUserCommand_BothFields(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotBody map[string]interface{}
	deps := updateUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "172.16.200.185", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotBody = body.(map[string]interface{})
			return []byte(`{"message":"User updated successfully"}`), nil
		},
	}

	var out bytes.Buffer
	err := runUpdateUserCommand([]string{"--profile", "prod", "--username", "test", "--new-password", "newpassword", "--new-role", "operator"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	want := map[string]interface{}{"newPassword": "newpassword", "newRole": "operator"}
	if !reflect.DeepEqual(gotBody, want) {
		t.Fatalf("unexpected body\nwant: %#v\ngot:  %#v", want, gotBody)
	}
	if !strings.Contains(out.String(), "User updated successfully") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunUpdateUserCommand_UsesEnvProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "env-profile")

	var gotProfile string
	deps := updateUserDeps{
		GetCreds: func(profile string) types.Credential {
			gotProfile = profile
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"message":"User updated successfully"}`), nil
		},
	}

	var out bytes.Buffer
	err := runUpdateUserCommand([]string{"--username", "test", "--new-role", "operator"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotProfile != "env-profile" {
		t.Fatalf("expected env profile, got %q", gotProfile)
	}
}

func TestRunUpdateUserCommand_MissingProfile(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := updateUserDeps{}
	var out bytes.Buffer
	err := runUpdateUserCommand([]string{"--username", "test", "--new-role", "operator"}, deps, &out)
	if !errors.Is(err, errUpdateUserUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunUpdateUserCommand_MissingUsername(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "p1")

	deps := updateUserDeps{GetCreds: func(profile string) types.Credential {
		return types.Credential{IP: "127.0.0.1", Port: "8443"}
	}}

	var out bytes.Buffer
	err := runUpdateUserCommand([]string{"--profile", "p1", "--new-role", "operator"}, deps, &out)
	if !errors.Is(err, errUpdateUserUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunUpdateUserCommand_NoUpdateFields(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "p1")

	deps := updateUserDeps{GetCreds: func(profile string) types.Credential {
		return types.Credential{IP: "127.0.0.1", Port: "8443"}
	}}

	var out bytes.Buffer
	err := runUpdateUserCommand([]string{"--profile", "p1", "--username", "test"}, deps, &out)
	if !errors.Is(err, errUpdateUserUsage) {
		t.Fatalf("expected usage error, got %v", err)
	}
	if !strings.Contains(out.String(), "No updates requested") {
		t.Fatalf("expected no updates message, got %q", out.String())
	}
}

func TestRunUpdateUserCommand_MissingIPPort(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := updateUserDeps{GetCreds: func(profile string) types.Credential {
		return types.Credential{}
	}}

	var out bytes.Buffer
	err := runUpdateUserCommand([]string{"--profile", "p1", "--username", "test", "--new-role", "operator"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "missing IP/Port credentials") {
		t.Fatalf("expected missing IP/Port error, got %v", err)
	}
}

func TestRunUpdateUserCommand_RequestError(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := updateUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	var out bytes.Buffer
	err := runUpdateUserCommand([]string{"--profile", "p1", "--username", "test", "--new-role", "operator"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "error updating user") {
		t.Fatalf("expected request error, got %v", err)
	}
}

func TestRunUpdateUserCommand_InvalidJSON(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := updateUserDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "127.0.0.1", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte("not-json"), nil
		},
	}

	var out bytes.Buffer
	err := runUpdateUserCommand([]string{"--profile", "p1", "--username", "test", "--new-role", "operator"}, deps, &out)
	if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
		t.Fatalf("expected parse error, got %v", err)
	}
}
