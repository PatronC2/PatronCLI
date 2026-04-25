package files

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"patroncli/common"
	"patroncli/types"
	"strings"
)

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
