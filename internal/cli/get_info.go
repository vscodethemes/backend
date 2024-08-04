package cli

import (
	"context"
	"encoding/json"
	"fmt"
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
	cmd := exec.CommandContext(ctx, "npx", "vscodethemes", "info", "-d", extensionPath)
	cmd.Dir = "cli"

	output, err := cmd.Output()
	if err != nil {
		// Read stderr to get the error message.
		fmt.Println("stderr: ", string(err.(*exec.ExitError).Stderr))
		return nil, fmt.Errorf("failed to get info: %w", err)
	}

	var result GetInfoResult
	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal output: %w", err)
	}

	return &result, nil
}
