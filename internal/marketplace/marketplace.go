package marketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vscodethemes/backend/internal/marketplace/qo"
)

type Client struct {
	BaseUrl    string
	httpClient *http.Client
}

type ClientOption func(*Client)

func WithBaseUrl(baseUrl string) ClientOption {
	return func(c *Client) {
		c.BaseUrl = baseUrl
	}
}

func NewClient(opts ...ClientOption) *Client {
	// Default client options
	client := &Client{
		BaseUrl:    "https://marketplace.visualstudio.com/_apis",
		httpClient: &http.Client{},
	}

	// Apply option overrides.
	for _, applyOpt := range opts {
		applyOpt(client)
	}

	return client
}

type QueryBody struct {
	AssetTypes *string           `json:"assetTypes"`
	Filters    []qo.QueryOptions `json:"filters"`
	Flags      int               `json:"flags"`
}

type QueryResponse struct {
	Results []QueryResult `json:"results"`
}

type QueryResult struct {
	Extensions []ExtensionResult `json:"extensions"`
}

type ExtensionResult struct {
	Publisher           ExtensionPublisherResult            `json:"publisher"`
	ExtensionID         string                              `json:"extensionId"`
	ExtensionName       string                              `json:"extensionName"`
	DisplayName         string                              `json:"displayName"`
	Flags               string                              `json:"flags"`
	LastUpdated         string                              `json:"lastUpdated"`
	PublishedDate       string                              `json:"publishedDate"`
	ReleaseDate         string                              `json:"releaseDate"`
	ShortDescription    *string                             `json:"shortDescription"`
	Versions            []ExtensionVersionResult            `json:"versions"`
	Categories          []string                            `json:"categories"`
	Tags                []string                            `json:"tags"`
	Stastistics         []ExtensionStatisticsResult         `json:"statistics"`
	InstallationTargets []ExtensionInstallationTargetResult `json:"installationTargets"`
	DeploymentType      int                                 `json:"deploymentType"`
}

type ExtensionPublisherResult struct {
	PublisherID   string `json:"publisherId"`
	PublisherName string `json:"publisherName"`
	DisplayName   string `json:"displayName"`
	Flags         string `json:"flags"`
}

type ExtensionVersionResult struct {
	Version     string    `json:"version"`
	Flags       string    `json:"flags"`
	LastUpdated time.Time `json:"lastUpdated"`
	Files       []struct {
		AssetType string `json:"assetType"`
		Source    string `json:"source"`
	} `json:"files"`
	Properties []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"properties"`
	AssetURI         string `json:"assetUri"`
	FallbackAssetURI string `json:"fallbackAssetUri"`
}

type ExtensionStatisticsResult struct {
	StatisticName string  `json:"statisticName"`
	Value         float64 `json:"value"`
}

type ExtensionInstallationTargetResult struct {
	Target        string `json:"target"`
	TargetVersion string `json:"targetVersion"`
}

func (e ExtensionResult) GetLatestVersion() *ExtensionVersionResult {
	var latestVersion *ExtensionVersionResult
	for _, version := range e.Versions {
		if latestVersion == nil {
			latestVersion = &version
			continue
		}

		if version.LastUpdated.After(latestVersion.LastUpdated) {
			latestVersion = &version
		}
	}

	return latestVersion
}

func (e ExtensionResult) GetPackageURL() string {
	latestVersion := e.GetLatestVersion()
	if latestVersion == nil {
		return ""
	}

	for _, file := range latestVersion.Files {
		if file.AssetType == "Microsoft.VisualStudio.Services.VSIXPackage" {
			return file.Source
		}
	}

	return ""
}

func (m Client) NewQuery(ctx context.Context, opts ...qo.QueryOption) ([]ExtensionResult, error) {
	// Default query options.
	queryOptions := &qo.QueryOptions{
		PageNumber: 1,
		PageSize:   100,
		SortBy:     qo.SortByLastUpdated,
		Direction:  qo.DirectionAsc,
		Criteria:   []qo.QueryOptionCriteria{},
	}

	// Apply option overrides.
	for _, applyOpt := range opts {
		applyOpt(queryOptions)
	}

	// Build the query body.
	reqBody := QueryBody{
		Filters: []qo.QueryOptions{*queryOptions},
		Flags:   870, // TODO: Do we need this?
	}

	reqJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query body: %w", err)
	}

	// Build the request.
	url := m.BaseUrl + "/public/gallery/extensionquery"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJson))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json;api-version=5.2-preview.1;excludeUrls=true")

	// Send the request.
	res, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	// Read the response body.
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var queryRes QueryResponse
	if err := json.Unmarshal(resBody, &queryRes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if len(queryRes.Results) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	return queryRes.Results[0].Extensions, nil
}
