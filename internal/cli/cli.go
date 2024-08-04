package cli

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// )

// type GenerateImagesResult struct {
// 	Extension Extension       `json:"extension"`
// 	Themes    []ThemeMetadata `json:"themes"`
// }

// type Extension struct {
// 	DisplayName string  `json:"displayName"`
// 	Description string  `json:"description"`
// 	GithubLink  *string `json:"githubLink"`
// }

// type ThemeMetadata struct {
// 	Theme  Theme   `json:"theme"`
// 	Images []Image `json:"images"`
// }

// type Theme struct {
// 	Path           string           `json:"path"`
// 	DisplayName    string           `json:"displayName"`
// 	Type           string           `json:"type"`
// 	Colors         Colors           `json:"colors"`
// 	LanguageTokens []LanguageTokens `json:"languageTokens"`
// }

// type Colors struct {
// 	EditorBackground              string  `json:"editorBackground"`
// 	EditorForeground              string  `json:"editorForeground"`
// 	ActivityBarBackground         string  `json:"activityBarBackground"`
// 	ActivityBarForeground         string  `json:"activityBarForeground"`
// 	ActivityBarInActiveForeground string  `json:"activityBarInActiveForeground"`
// 	ActivityBarBorder             *string `json:"activityBarBorder"`
// 	ActivityBarActiveBorder       string  `json:"activityBarActiveBorder"`
// 	ActivityBarActiveBackground   *string `json:"activityBarActiveBackground"`
// 	ActivityBarBadgeBackground    string  `json:"activityBarBadgeBackground"`
// 	ActivityBarBadgeForeground    string  `json:"activityBarBadgeForeground"`
// 	TabsContainerBackground       *string `json:"tabsContainerBackground"`
// 	TabsContainerBorder           *string `json:"tabsContainerBorder"`
// 	StatusBarBackground           *string `json:"statusBarBackground"`
// 	StatusBarForeground           string  `json:"statusBarForeground"`
// 	StatusBarBorder               *string `json:"statusBarBorder"`
// 	TabActiveBackground           *string `json:"tabActiveBackground"`
// 	TabInactiveBackground         *string `json:"tabInactiveBackground"`
// 	TabActiveForeground           string  `json:"tabActiveForeground"`
// 	TabBorder                     string  `json:"tabBorder"`
// 	TabActiveBorder               *string `json:"tabActiveBorder"`
// 	TabActiveBorderTop            *string `json:"tabActiveBorderTop"`
// 	TitleBarActiveBackground      string  `json:"titleBarActiveBackground"`
// 	TitleBarActiveForeground      string  `json:"titleBarActiveForeground"`
// 	TitleBarBorder                *string `json:"titleBarBorder"`
// }

// type LanguageTokens struct {
// 	Language string    `json:"language"`
// 	Tokens   [][]Token `json:"tokens"`
// }

// type Token struct {
// 	Text  string `json:"text"`
// 	Style Style  `json:"style"`
// }

// type Style struct {
// 	Color          *string `json:"color"`
// 	FontWeight     *string `json:"fontWeight"`
// 	FontStyle      *string `json:"fontStyle"`
// 	TextDecoration *string `json:"textDecoration"`
// }

// type Image struct {
// 	Language string            `json:"language"`
// 	Paths    map[string]string `json:"paths"`
// }

// func GenerateImages(extensionPath string) (*GenerateImagesResult, error) {
// 	extensionPathAbs, err := filepath.Abs(extensionPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get absolute path: %w", err)
// 	}

// 	outputPath := fmt.Sprintf("%s/images", extensionPathAbs)
// 	cmd := exec.Command("npx", "vscodethemes", "images", "-d", extensionPathAbs, "-o", outputPath)
// 	cmd.Dir = "cli"

// 	_, err = cmd.Output()
// 	if err != nil {
// 		// Read stderr to get the error message.
// 		fmt.Println("stderr: ", string(err.(*exec.ExitError).Stderr))
// 		return nil, fmt.Errorf("failed to generate images: %w", err)
// 	}

// 	// Read output file as JSON.
// 	outputFilePath := fmt.Sprintf("%s/output.json", outputPath)
// 	outputFile, err := os.ReadFile(outputFilePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read output file: %w", err)
// 	}

// 	var result GenerateImagesResult
// 	err = json.Unmarshal(outputFile, &result)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal output file: %w", err)
// 	}

// 	return &result, nil
// }
