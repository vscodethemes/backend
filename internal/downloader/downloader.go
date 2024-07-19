package downloader

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Downloader struct {
	PackagePath string
	ExtractDir  string
}

func New(dir, extensionSlug string) *Downloader {
	return &Downloader{
		PackagePath: path.Join(dir, fmt.Sprintf("%s.VSIXPackage", extensionSlug)),
		ExtractDir:  path.Join(dir, extensionSlug),
	}
}

func (d *Downloader) Download(ctx context.Context, url string) error {
	file, err := os.Create(d.PackagePath)
	if err != nil {
		return fmt.Errorf("failed to create package file: %w", err)
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write package file: %w", err)
	}

	return nil
}

func (d *Downloader) Extract() error {
	reader, err := zip.OpenReader(d.PackagePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	err = os.MkdirAll(d.ExtractDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}

	// Function to extract and write a file from a zip. This is called for
	// each file in the zip and checks for ZipSlip. A local function is used
	// instead of adding directly to the for loop below so that we can defer
	// the closing of the file.
	extractAndWriteFile := func(zipFile *zip.File) error {
		readCloser, err := zipFile.Open()
		if err != nil {
			return err
		}
		defer readCloser.Close()

		path := filepath.Join(d.ExtractDir, zipFile.Name)

		// Check for ZipSlip (Directory traversal).
		if !strings.HasPrefix(path, filepath.Clean(d.ExtractDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if zipFile.FileInfo().IsDir() {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			file, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer file.Close()

			_, err = io.Copy(file, readCloser)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Extract all files.
	for _, file := range reader.File {
		err := extractAndWriteFile(file)
		if err != nil {
			return err
		}
	}

	return nil
}
