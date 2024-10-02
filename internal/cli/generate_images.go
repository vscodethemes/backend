package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type GenerateImagesResult struct {
	Theme     Theme            `json:"theme"`
	Languages []LanguageResult `json:"languages"`
}

type Theme struct {
	Path        string `json:"path"`
	DisplayName string `json:"displayName"`
	Type        string `json:"type"`
	Colors      Colors `json:"colors"`
}

type Colors struct {
	EditorBackground              string  `json:"editorBackground"`
	EditorForeground              string  `json:"editorForeground"`
	ActivityBarBackground         string  `json:"activityBarBackground"`
	ActivityBarForeground         string  `json:"activityBarForeground"`
	ActivityBarInActiveForeground string  `json:"activityBarInActiveForeground"`
	ActivityBarBorder             *string `json:"activityBarBorder"`
	ActivityBarActiveBorder       string  `json:"activityBarActiveBorder"`
	ActivityBarActiveBackground   *string `json:"activityBarActiveBackground"`
	ActivityBarBadgeBackground    string  `json:"activityBarBadgeBackground"`
	ActivityBarBadgeForeground    string  `json:"activityBarBadgeForeground"`
	TabsContainerBackground       *string `json:"tabsContainerBackground"`
	TabsContainerBorder           *string `json:"tabsContainerBorder"`
	StatusBarBackground           *string `json:"statusBarBackground"`
	StatusBarForeground           string  `json:"statusBarForeground"`
	StatusBarBorder               *string `json:"statusBarBorder"`
	TabActiveBackground           *string `json:"tabActiveBackground"`
	TabInactiveBackground         *string `json:"tabInactiveBackground"`
	TabActiveForeground           string  `json:"tabActiveForeground"`
	TabBorder                     string  `json:"tabBorder"`
	TabActiveBorder               *string `json:"tabActiveBorder"`
	TabActiveBorderTop            *string `json:"tabActiveBorderTop"`
	TitleBarActiveBackground      string  `json:"titleBarActiveBackground"`
	TitleBarActiveForeground      string  `json:"titleBarActiveForeground"`
	TitleBarBorder                *string `json:"titleBarBorder"`
}

type LanguageResult struct {
	Language Language  `json:"language"`
	Tokens   [][]Token `json:"tokens"`
	SvgPath  string    `json:"svgPath"`
	PngPath  string    `json:"pngPath"`
}

type Language struct {
	Name      string `json:"name"`
	ExtName   string `json:"extName"`
	ScopeName string `json:"scopeName"`
	Grammar   string `json:"grammar"`
	Template  string `json:"template"`
	TabName   string `json:"tabName"`
}

type Token struct {
	Text  string `json:"text"`
	Style Style  `json:"style"`
}

type Style struct {
	Color          *string `json:"color"`
	FontWeight     *string `json:"fontWeight"`
	FontStyle      *string `json:"fontStyle"`
	TextDecoration *string `json:"textDecoration"`
}

func GenerateImages(ctx context.Context, extensionPath string, theme ThemeContribute, outputDir string) (*GenerateImagesResult, error) {
	// Ensure the extension directory exists.
	if _, err := os.Stat(extensionPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("extension directory does not exist: %w", err)
	}

	args := []string{"vscodethemes", "images"}
	args = append(args, "--dir", extensionPath)
	args = append(args, "--uiTheme", theme.UITheme)
	args = append(args, "--path", theme.Path)
	args = append(args, "--output", outputDir)
	if theme.Label != nil {
		args = append(args, "--label", *theme.Label)
	}

	cmd := exec.CommandContext(ctx, "npx", args...)
	cmd.Dir = "cli"

	output, err := cmd.Output()
	if err != nil {
		stderr := string(err.(*exec.ExitError).Stderr)
		return nil, fmt.Errorf("failed to generate images: %s", stderr)
	}

	var result GenerateImagesResult
	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal output: %w", err)
	}

	return &result, nil
}
