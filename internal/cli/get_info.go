package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type GetInfoResult struct {
	Extension        Extension         `json:"extension"`
	ThemeContributes []ThemeContribute `json:"themeContributes"`
}

type Extension struct {
	DisplayName string  `json:"displayName"`
	Description string  `json:"description"`
	GithubLink  *string `json:"githubLink"`
}

type ThemeContribute struct {
	Path    string  `json:"path"`
	UITheme string  `json:"uiTheme"`
	Label   *string `json:"label"`
}

func GetInfo(ctx context.Context, extensionPath string) (*GetInfoResult, error) {
	// Ensure the extension directory exists.
	if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("extension directory does not exist: %w", err)
	}

	cmd := exec.CommandContext(ctx, "npx", "vscodethemes", "info", "-d", extensionPath)
	// cmd := exec.CommandContext(ctx, "npx", "vscodethemes", "help")
	// TODO: Make cli path configurable?
	// cmd.Dir = "/cli"
	cmd.Dir = "cli"

	output, err := cmd.Output()
	if err != nil {
		stderr := string(err.(*exec.ExitError).Stderr)
		return nil, fmt.Errorf("failed to get info: %s", stderr)
	}

	var result GetInfoResult
	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal output: %w", err)
	}

	return &result, nil
}
