package gravatar

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	authorizationEndpoint = "https://public-api.wordpress.com/oauth2/authorize"
	tokenEndpoint         = "https://public-api.wordpress.com/oauth2/token"
)

// OAuthConfig holds OAuth 2.0 configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	state        string
	codeChan     chan string
	errChan      chan error
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	BlogID      string `json:"blog_id"`
	BlogURL     string `json:"blog_url"`
}

// NewOAuthConfig() - creates a new OAuth configuration
/* clientID (string) - client ID for OAuth authentication
   clientSecret (string) - client secret for OAuth authentication
   redirectURI (string) - redirect URI for OAuth authentication */
func NewOAuthConfig(clientID, clientSecret, redirectURI string) *OAuthConfig {
	return &OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		Scopes:       []string{"auth", "gravatar-profile:manage"},
		codeChan:     make(chan string),
		errChan:      make(chan error),
	}
}

// StartOAuthFlow() - initiates the OAuth flow and returns an access token
// verbose (bool) - enable verbose logging
func (c *OAuthConfig) StartOAuthFlow(verbose bool) (string, error) {
	// generate random state for CSRF protection
	c.state = generateRandomState()

	// start local server
	server := &http.Server{Addr: ":8080"}
	http.HandleFunc("/callback", c.callbackHandler)

	// start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			c.errChan <- fmt.Errorf("failed to start callback server: %v", err)
		}
	}()
	
	time.Sleep(100 * time.Millisecond)

	// build authorization URL
	authURL := c.buildAuthURL()

	if verbose {
		fmt.Println("Opening browser for authorization...")
		fmt.Printf("If browser doesn't open, visit: %s\n", authURL)
	}

	// open browser
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser automatically: %v\n", err)
		fmt.Printf("Please visit this URL manually: %s\n", authURL)
	}

	// wait for callback or timeout
	var code string
	select {
	case code = <-c.codeChan:
		// got authorization code
	case err := <-c.errChan:
		server.Shutdown(context.Background())
		return "", err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return "", fmt.Errorf("authorization timeout after 5 minutes")
	}

	// shutdown server
	server.Shutdown(context.Background())

	if verbose {
		fmt.Println("Authorization successful! Exchanging code for token...")
	}

	// exchange code for token
	token, err := c.exchangeCodeForToken(code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code for token: %v", err)
	}

	return token, nil
}

// buildAuthURL() - constructs the OAuth authorization URL
//c (*OAuthConfig) - OAuth configuration
func (c *OAuthConfig) buildAuthURL() string {
	params := url.Values{}
	params.Set("client_id", c.ClientID)
	params.Set("redirect_uri", c.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(c.Scopes, " "))
	params.Set("state", c.state)

	return authorizationEndpoint + "?" + params.Encode()
}

// callbackHandler() - handles the OAuth callback
/* w (*http.ResponseWriter) - HTTP response writer
   r (*http.Request) - HTTP request */
func (c *OAuthConfig) callbackHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// check for errors
	if errMsg := query.Get("error"); errMsg != "" {
		c.errChan <- fmt.Errorf("authorization denied: %s", errMsg)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "<html><body><h1>Authorization Denied</h1><p>%s</p><p>You can close this window.</p></body></html>", errMsg)
		return
	}

	// verify state
	state := query.Get("state")
	if state != c.state {
		c.errChan <- fmt.Errorf("invalid state parameter (CSRF protection)")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "<html><body><h1>Error</h1><p>Invalid state parameter</p><p>You can close this window.</p></body></html>")
		return
	}

	// get authorization code
	code := query.Get("code")
	if code == "" {
		c.errChan <- fmt.Errorf("no authorization code received")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "<html><body><h1>Error</h1><p>No authorization code received</p><p>You can close this window.</p></body></html>")
		return
	}

	// send success response to user
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<html><body><h1>Success!</h1><p>Authorization successful. You can close this window and return to the terminal.</p></body></html>")

	// send code to main flow
	c.codeChan <- code
}

// exchangeCodeForToken() - exchanges the authorization code for an access token
// code (string) - authorization code to exchange for access token
func (c *OAuthConfig) exchangeCodeForToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", c.RedirectURI)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %v", err)
	}

	return tokenResp.AccessToken, nil
}

// openBrowser() - opens the default browser to the specified URL
// url (string) - URL to open in the browser
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}

// generateRandomState() - generates a random state string for CSRF protection
func generateRandomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
