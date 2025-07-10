package database

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/registry/internal/model"
)

// ReadSeedFile reads and parses the seed.json file - exported for use by all database implementations
// Supports both local file paths and HTTP URLs
func ReadSeedFile(path string) ([]model.ServerDetail, error) {
	log.Printf("Reading seed file from %s", path)

	// Set default seed file path if not provided
	if path == "" {
		// Try to find the seed.json in the data directory
		path = filepath.Join("data", "seed.json")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("seed file not found at %s", path)
		}
	}

	var fileContent []byte
	var err error

	// Check if path is an HTTP URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// Read from HTTP URL
		fileContent, err = readFromHTTP(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read from HTTP URL: %w", err)
		}
	} else {
		// Read from local file
		fileContent, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
	}

	// Parse the JSON content
	var servers []model.ServerDetail
	if err := json.Unmarshal(fileContent, &servers); err != nil {
		// Try parsing as a raw JSON array and then convert to our model
		var rawData []map[string]interface{}
		if jsonErr := json.Unmarshal(fileContent, &rawData); jsonErr != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w (original error: %w)", jsonErr, err)
		}
	}

	log.Printf("Found %d server entries in seed file", len(servers))
	return servers, nil
}

// readFromHTTP reads content from an HTTP URL with timeout
func readFromHTTP(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
