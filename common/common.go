package common

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"patroncli/config"
	"patroncli/types"
	"time"

	"golang.org/x/net/proxy"
)

func buildHTTPClient(profile types.Credential) (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Ignore self-signed certificates
		},
	}

	// If SOCKS5 is configured, tunnel outbound connections through it.
	if profile.SOCKS5Enabled {
		if profile.SOCKS5Host == "" || profile.SOCKS5Port == "" {
			return nil, fmt.Errorf("SOCKS5 proxy enabled but host/port not set")
		}

		proxyAddr := net.JoinHostPort(profile.SOCKS5Host, profile.SOCKS5Port)

		var auth *proxy.Auth
		if profile.SOCKS5Username != "" {
			auth = &proxy.Auth{
				User:     profile.SOCKS5Username,
				Password: profile.SOCKS5Password,
			}
		}

		dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}

		// Prefer DialContext if supported; otherwise fallback to Dial (no ctx cancel).
		type contextDialer interface {
			DialContext(ctx context.Context, network, addr string) (net.Conn, error)
		}

		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			if cd, ok := dialer.(contextDialer); ok {
				return cd.DialContext(ctx, network, addr)
			}
			return dialer.Dial(network, addr)
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}

// MakeRequest is a generic function for API requests (GET, POST, PUT, DELETE).
func MakeRequest(method, url string, profile types.Credential, body interface{}) ([]byte, error) {
	client, err := buildHTTPClient(profile)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", profile.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return responseData, nil
}

// MakeMultipartRequest sends a multipart/form-data request with fields and an optional file part.
func MakeMultipartRequest(method, url string, profile types.Credential, fields map[string]string, fileField, fileName string, fileContent []byte) ([]byte, error) {
	client, err := buildHTTPClient(profile)
	if err != nil {
		return nil, err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("failed to write multipart field %q: %w", k, err)
		}
	}

	if fileField != "" {
		part, err := writer.CreateFormFile(fileField, fileName)
		if err != nil {
			return nil, fmt.Errorf("failed to create multipart file part: %w", err)
		}
		if _, err := part.Write(fileContent); err != nil {
			return nil, fmt.Errorf("failed to write multipart file content: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize multipart body: %w", err)
	}

	req, err := http.NewRequest(method, url, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", profile.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed: status=%d body=%s", resp.StatusCode, string(b))
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	return responseData, nil
}

func GetCreds(profileName string) types.Credential {
	credentialsPath := config.GetCredentialsPath()
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		fmt.Printf("Warning: failed to read credentials file: %v\n", err)
		return types.Credential{}
	}

	var creds []types.Credential
	err = json.Unmarshal(data, &creds)
	if err != nil {
		return types.Credential{}
	}

	for _, cred := range creds {
		if cred.Profile == profileName {
			return cred
		}
	}
	return types.Credential{}
}
