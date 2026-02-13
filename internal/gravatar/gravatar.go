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

// NewClient() - creates a new Gravatar client with OAuth credentials
/* clientID (string) - client ID for OAuth authentication
   clientSecret (string) - client secret for OAuth authentication
   redirectURI (string) - redirect URI for OAuth authentication
   verbose (bool) - enable verbose logging */
func NewClient(clientID, clientSecret, redirectURI string, verbose bool) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		Verbose:      verbose,
	}
}

// Authenticate() - performs OAuth flow and obtains an access token
// c (*Client) - Gravatar client to authenticate
func (c *Client) Authenticate() error {
	oauth := NewOAuthConfig(c.ClientID, c.ClientSecret, c.RedirectURI)

	token, err := oauth.StartOAuthFlow(c.Verbose)
	if err != nil {
		return err
	}

	c.AccessToken = token
	return nil
}

// UploadAvatar() - uploads an image to Gravatar using the REST API
/* c (*Client) - Gravatar client to upload avatar
   imagePath (string) - path to the image to upload */
func (c *Client) UploadAvatar(imagePath string) error {
	if c.AccessToken == "" {
		return fmt.Errorf("not authenticated - call Authenticate() first")
	}

	// Gravatar requires square images - crop if necessary
	if c.Verbose {
		fmt.Println("Checking if image needs to be cropped to square...")
	}

	squareImagePath, err := cropToSquare(imagePath)
	if err != nil {
		return fmt.Errorf("failed to crop image to square: %v", err)
	}

	if c.Verbose {
		if squareImagePath != imagePath {
			fmt.Printf("Cropped image to square: %s\n", squareImagePath)
		} else {
			fmt.Println("Image is already square, no cropping needed")
		}
	}

	// clean up temporary square image if it's different from original
	if squareImagePath != imagePath {
		defer os.Remove(squareImagePath)
	}

	file, err := os.Open(squareImagePath)
	if err != nil {
		return fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	// create multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("image", filepath.Base(squareImagePath))
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

	// create the HTTP request with select_avatar parameter
	uploadURL := apiBaseURL + "/me/avatars?select_avatar=true"
	req, err := http.NewRequest("POST", uploadURL, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// set headers
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
