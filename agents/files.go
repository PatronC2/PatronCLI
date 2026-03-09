package agents

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"patroncli/common"
	"patroncli/types"
	"strings"
)

func ListFilesCommand(args []string) {
	cmd := flag.NewFlagSet("list-files", flag.ExitOnError)
	profileName := cmd.String("profile", "", "The profile name to use")
	agentID := cmd.String("agent-id", "", "The agent UUID to list files for")
	query := cmd.String("query", "", "Comma-separated fields to include in output")
	cmd.Parse(args)

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}
	if selectedProfile == "" {
		fmt.Println("No profile specified. Use --profile flag or set the PATRON_PROFILE environment variable.")
		os.Exit(1)
	}
	if strings.TrimSpace(*agentID) == "" {
		fmt.Println("No agent specified. Use --agent-id flag.")
		cmd.Usage()
		os.Exit(1)
	}

	profile := common.GetCreds(selectedProfile)
	if err := listFiles(profile, strings.TrimSpace(*agentID), strings.TrimSpace(*query)); err != nil {
		fmt.Println("Error listing files:", err)
		os.Exit(1)
	}
}

func listFiles(profile types.Credential, agentID, query string) error {
	url := fmt.Sprintf("https://%s:%s/api/files/list/%s", profile.IP, profile.Port, agentID)
	body, err := common.MakeRequest("GET", url, profile, nil)
	if err != nil {
		return fmt.Errorf("error listing files: %w", err)
	}

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	items := response.Data
	if query != "" {
		items = common.QueryFields(items, query)
	}

	output, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize files to JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func DownloadFileCommand(args []string) {
	cmd := flag.NewFlagSet("download-file", flag.ExitOnError)
	profileName := cmd.String("profile", "", "The profile name to use")
	fileID := cmd.String("file-id", "", "The file ID to download")
	outputPath := cmd.String("output", "", "Output file path on local disk")
	cmd.Parse(args)

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}
	if selectedProfile == "" {
		fmt.Println("No profile specified. Use --profile flag or set the PATRON_PROFILE environment variable.")
		os.Exit(1)
	}
	if strings.TrimSpace(*fileID) == "" {
		fmt.Println("No file specified. Use --file-id flag.")
		cmd.Usage()
		os.Exit(1)
	}
	if strings.TrimSpace(*outputPath) == "" {
		fmt.Println("No output path specified. Use --output flag.")
		cmd.Usage()
		os.Exit(1)
	}

	profile := common.GetCreds(selectedProfile)
	if err := downloadFile(profile, strings.TrimSpace(*fileID), strings.TrimSpace(*outputPath)); err != nil {
		fmt.Println("Error downloading file:", err)
		os.Exit(1)
	}
}

func downloadFile(profile types.Credential, fileID, outputPath string) error {
	url := fmt.Sprintf("https://%s:%s/api/files/download/%s", profile.IP, profile.Port, fileID)
	body, err := common.MakeRequest("GET", url, profile, nil)
	if err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	if err := os.WriteFile(outputPath, body, 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Saved file to %s\n", outputPath)
	return nil
}

func UploadFileCommand(args []string) {
	cmd := flag.NewFlagSet("upload-file", flag.ExitOnError)
	profileName := cmd.String("profile", "", "The profile name to use")
	agentID := cmd.String("agent-id", "", "The agent UUID")
	path := cmd.String("path", "", "Destination/source path on target")
	transferType := cmd.String("transfer-type", "", "Optional transfer type override (Upload or Download)")
	sourcePath := cmd.String("source", "", "Local file path. If set, transfer-type defaults to Download; if not set, defaults to Upload.")
	cmd.Parse(args)

	selectedProfile := os.Getenv("PATRON_PROFILE")
	if *profileName != "" {
		selectedProfile = *profileName
	}
	if selectedProfile == "" {
		fmt.Println("No profile specified. Use --profile flag or set the PATRON_PROFILE environment variable.")
		os.Exit(1)
	}
	if strings.TrimSpace(*agentID) == "" {
		fmt.Println("No agent specified. Use --agent-id flag.")
		cmd.Usage()
		os.Exit(1)
	}
	if strings.TrimSpace(*path) == "" {
		fmt.Println("No path specified. Use --path flag.")
		cmd.Usage()
		os.Exit(1)
	}
	profile := common.GetCreds(selectedProfile)
	if err := uploadFile(
		profile,
		strings.TrimSpace(*agentID),
		strings.TrimSpace(*path),
		strings.TrimSpace(*sourcePath),
		strings.TrimSpace(*transferType),
	); err != nil {
		fmt.Println("Error uploading file:", err)
		os.Exit(1)
	}
}

func uploadFile(profile types.Credential, agentID, targetPath, sourcePath, transferTypeOverride string) error {
	transferType, err := resolveTransferType(sourcePath, transferTypeOverride)
	if err != nil {
		return err
	}

	fields := map[string]string{
		"uuid":         agentID,
		"path":         targetPath,
		"transfertype": transferType,
	}

	url := fmt.Sprintf("https://%s:%s/api/files/upload", profile.IP, profile.Port)
	if transferType == "Upload" {
		body, err := common.MakeMultipartRequest("POST", url, profile, fields, "", "", nil)
		if err != nil {
			return fmt.Errorf("error uploading file: %w", err)
		}
		return printUploadResponse(body)
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	body, err := common.MakeMultipartRequest("POST", url, profile, fields, "file", filepath.Base(sourcePath), content)
	if err != nil {
		return fmt.Errorf("error uploading file: %w", err)
	}
	return printUploadResponse(body)
}

func printUploadResponse(body []byte) error {
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	output, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize response to JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func resolveTransferType(sourcePath, override string) (string, error) {
	normalizedOverride := strings.TrimSpace(override)
	if normalizedOverride != "" {
		switch strings.ToLower(normalizedOverride) {
		case "upload":
			if strings.TrimSpace(sourcePath) != "" {
				return "", fmt.Errorf("--source cannot be used with transfer-type Upload")
			}
			return "Upload", nil
		case "download":
			if strings.TrimSpace(sourcePath) == "" {
				return "", fmt.Errorf("--source is required when transfer-type is Download")
			}
			return "Download", nil
		default:
			return "", fmt.Errorf("invalid --transfer-type %q (expected Upload or Download)", override)
		}
	}

	if strings.TrimSpace(sourcePath) != "" {
		return "Download", nil
	}
	return "Upload", nil
}
