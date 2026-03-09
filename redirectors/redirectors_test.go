package redirectors

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"patroncli/types"
)

func TestRunGetRedirectorsCommand_Success(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotMethod, gotURL string
	deps := redirectorDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "172.16.200.185", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotMethod, gotURL = method, url
			return []byte(`{"data":[{"id":"r1","name":"edge-a"}]}`), nil
		},
	}

	var out bytes.Buffer
	err := runGetRedirectorsCommand([]string{"--profile", "prod"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotMethod != "GET" || gotURL != "https://172.16.200.185:8443/api/redirectors" {
		t.Fatalf("unexpected request %s %s", gotMethod, gotURL)
	}
	if !strings.Contains(out.String(), "edge-a") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunGetRedirectorsCommand_FilterAndQuery(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	deps := redirectorDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "172.16.200.185", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			return []byte(`{"data":[{"id":"r1","name":"edge-a","forward_ip":"10.0.0.2"},{"id":"r2","name":"edge-b","forward_ip":"10.0.0.3"}]}`), nil
		},
	}

	var out bytes.Buffer
	err := runGetRedirectorsCommand([]string{"--profile", "prod", "--filter", "name=edge-a", "--query", "name, forward_ip"}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `"name": "edge-a"`) {
		t.Fatalf("expected filtered redirector in output, got %q", output)
	}
	if strings.Contains(output, `"name": "edge-b"`) {
		t.Fatalf("did not expect non-matching redirector, got %q", output)
	}
	if strings.Contains(output, `"id":`) {
		t.Fatalf("did not expect id field when query is applied, got %q", output)
	}
}

func TestRunGetRedirectorsCommand_Errors(t *testing.T) {
	t.Run("missing profile", func(t *testing.T) {
		t.Setenv("PATRON_PROFILE", "")
		err := runGetRedirectorsCommand(nil, redirectorDeps{}, &bytes.Buffer{})
		if !errors.Is(err, errGetRedirectorsUsage) {
			t.Fatalf("expected usage error, got %v", err)
		}
	})

	t.Run("request error", func(t *testing.T) {
		t.Setenv("PATRON_PROFILE", "")
		deps := redirectorDeps{
			GetCreds: func(profile string) types.Credential { return types.Credential{IP: "127.0.0.1", Port: "8443"} },
			MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
				return nil, errors.New("boom")
			},
		}
		err := runGetRedirectorsCommand([]string{"--profile", "p1"}, deps, &bytes.Buffer{})
		if err == nil || !strings.Contains(err.Error(), "error fetching redirectors") {
			t.Fatalf("expected request error, got %v", err)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Setenv("PATRON_PROFILE", "")
		deps := redirectorDeps{
			GetCreds: func(profile string) types.Credential { return types.Credential{IP: "127.0.0.1", Port: "8443"} },
			MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
				return []byte("nope"), nil
			},
		}
		err := runGetRedirectorsCommand([]string{"--profile", "p1"}, deps, &bytes.Buffer{})
		if err == nil || !strings.Contains(err.Error(), "failed to parse response") {
			t.Fatalf("expected parse error, got %v", err)
		}
	})
}

func TestRunCreateRedirectorCommand_Success(t *testing.T) {
	t.Setenv("PATRON_PROFILE", "")

	var gotMethod, gotURL string
	var gotBody map[string]string
	deps := redirectorDeps{
		GetCreds: func(profile string) types.Credential {
			return types.Credential{IP: "172.16.200.185", Port: "8443"}
		},
		MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
			gotMethod, gotURL = method, url
			cast, ok := body.(map[string]string)
			if !ok {
				t.Fatalf("unexpected body type %T", body)
			}
			gotBody = cast
			return []byte("#!/bin/sh\necho ok"), nil
		},
	}

	var out bytes.Buffer
	err := runCreateRedirectorCommand([]string{
		"--profile", "prod",
		"--name", "edge-a",
		"--description", "primary edge",
		"--forward-ip", "10.0.0.2",
		"--forward-port", "8443",
		"--listen-ipv4", "192.168.1.2",
		"--listen-ipv6", "::1",
		"--listen-port", "443",
	}, deps, &out)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if gotMethod != "POST" || gotURL != "https://172.16.200.185:8443/api/redirector" {
		t.Fatalf("unexpected request %s %s", gotMethod, gotURL)
	}
	wantBody := map[string]string{
		"Name":        "edge-a",
		"Description": "primary edge",
		"ForwardIP":   "10.0.0.2",
		"ForwardPort": "8443",
		"ListenIPv4":  "192.168.1.2",
		"ListenIPv6":  "::1",
		"ListenPort":  "443",
	}
	if !reflect.DeepEqual(gotBody, wantBody) {
		t.Fatalf("unexpected body\nwant: %#v\ngot:  %#v", wantBody, gotBody)
	}
	if !strings.Contains(out.String(), "#!/bin/sh") {
		t.Fatalf("expected script output, got %q", out.String())
	}
}

func TestRunCreateRedirectorCommand_Errors(t *testing.T) {
	t.Run("missing required fields", func(t *testing.T) {
		t.Setenv("PATRON_PROFILE", "p1")
		deps := redirectorDeps{GetCreds: func(profile string) types.Credential { return types.Credential{IP: "127.0.0.1", Port: "8443"} }}
		var out bytes.Buffer
		err := runCreateRedirectorCommand([]string{"--profile", "p1", "--name", "edge"}, deps, &out)
		if !errors.Is(err, errCreateRedirectorUsage) {
			t.Fatalf("expected usage error, got %v", err)
		}
	})

	t.Run("request error", func(t *testing.T) {
		t.Setenv("PATRON_PROFILE", "")
		deps := redirectorDeps{
			GetCreds: func(profile string) types.Credential { return types.Credential{IP: "127.0.0.1", Port: "8443"} },
			MakeRequest: func(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
				return nil, errors.New("boom")
			},
		}
		err := runCreateRedirectorCommand([]string{"--profile", "p1", "--name", "edge", "--forward-ip", "10.0.0.1", "--forward-port", "8443", "--listen-ipv4", "0.0.0.0", "--listen-port", "443"}, deps, &bytes.Buffer{})
		if err == nil || !strings.Contains(err.Error(), "error creating redirector") {
			t.Fatalf("expected create error, got %v", err)
		}
	})
}
