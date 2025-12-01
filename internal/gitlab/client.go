package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a GitLab API client
type Client struct {
	baseURL     string
	accessToken string
	httpClient  *http.Client
}

// NewClient creates a new GitLab API client
func NewClient(baseURL, accessToken string) *Client {
	return &Client{
		baseURL:     baseURL,
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetMRDiff retrieves the diff content of a merge request
func (c *Client) GetMRDiff(projectID, mrIID int) (string, error) {
	// GitLab API endpoint: GET /api/v4/projects/:id/merge_requests/:merge_request_iid/changes
	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/changes", c.baseURL, projectID, mrIID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header
	req.Header.Set("PRIVATE-TOKEN", c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitLab API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var mrChanges struct {
		Changes []struct {
			OldPath string `json:"old_path"`
			NewPath string `json:"new_path"`
			Diff    string `json:"diff"`
		} `json:"changes"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &mrChanges); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Concatenate all diffs
	var fullDiff bytes.Buffer
	for _, change := range mrChanges.Changes {
		if change.Diff != "" {
			fullDiff.WriteString(fmt.Sprintf("--- a/%s\n+++ b/%s\n", change.OldPath, change.NewPath))
			fullDiff.WriteString(change.Diff)
			fullDiff.WriteString("\n")
		}
	}

	if fullDiff.Len() == 0 {
		return "", fmt.Errorf("no diff content found in merge request")
	}

	return fullDiff.String(), nil
}

// PostMRComment posts a comment to a merge request
func (c *Client) PostMRComment(projectID, mrIID int, comment string) error {
	// GitLab API endpoint: POST /api/v4/projects/:id/merge_requests/:merge_request_iid/notes
	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/notes", c.baseURL, projectID, mrIID)

	// Create request payload
	payload := map[string]string{
		"body": comment,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal comment payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header
	req.Header.Set("PRIVATE-TOKEN", c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitLab API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// TestConnection tests the GitLab API connection
func (c *Client) TestConnection() error {
	// GitLab API endpoint: GET /api/v4/user
	url := fmt.Sprintf("%s/api/v4/user", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitLab API authentication failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
