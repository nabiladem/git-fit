package gravatar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

var apiBaseURL = "https://api.gravatar.com/v3"

// Client handles Gravatar REST API interactions with OAuth
type Client struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AccessToken  string
	Verbose      bool
}

// NewClient creates a new Gravatar client with OAuth credentials
func NewClient(clientID, clientSecret, redirectURI string, verbose bool) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		Verbose:      verbose,
	}
}

// Authenticate performs OAuth flow and obtains an access token
func (c *Client) Authenticate() error {
	oauth := NewOAuthConfig(c.ClientID, c.ClientSecret, c.RedirectURI)

	token, err := oauth.StartOAuthFlow(c.Verbose)
	if err != nil {
		return err
	}

	c.AccessToken = token
	return nil
}

// UploadAvatar uploads an image to Gravatar using the REST API
func (c *Client) UploadAvatar(imagePath string) error {
	if c.AccessToken == "" {
		return fmt.Errorf("not authenticated - call Authenticate() first")
	}

	// Gravatar requires square images - crop if necessary
	squareImagePath, err := cropToSquare(imagePath)
	if err != nil {
		return fmt.Errorf("failed to crop image to square: %v", err)
	}

	// Clean up temporary square image if it's different from original
	if squareImagePath != imagePath {
		defer os.Remove(squareImagePath)
	}

	// Open the image file
	file, err := os.Open(squareImagePath)
	if err != nil {
		return fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	// Create multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the file to the form with field name "image"
	part, err := writer.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close multipart writer: %v", err)
	}

	// Create the HTTP request with select_avatar parameter
	uploadURL := apiBaseURL + "/me/avatars?select_avatar=true"
	req, err := http.NewRequest("POST", uploadURL, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
