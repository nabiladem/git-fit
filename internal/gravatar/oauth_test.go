package gravatar

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestBuildAuthURL() - test the buildAuthURL method
func TestBuildAuthURL(t *testing.T) {
	config := NewOAuthConfig("client-id", "client-secret", "http://localhost:8080/callback")
	config.state = "test-state"

	authURL := config.buildAuthURL()
	u, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("failed to parse auth URL: %v", err)
	}

	q := u.Query()
	if q.Get("client_id") != "client-id" {
		t.Errorf("expected client_id=client-id, got %s", q.Get("client_id"))
	}

	if q.Get("redirect_uri") != "http://localhost:8080/callback" {
		t.Errorf("expected redirect_uri=http://localhost:8080/callback, got %s", q.Get("redirect_uri"))
	}

	if q.Get("state") != "test-state" {
		t.Errorf("expected state=test-state, got %s", q.Get("state"))
	}

	if !strings.Contains(q.Get("scope"), "auth") {
		t.Error("scope missing 'auth'")
	}
}

// TestExchangeCodeForToken() - test the exchangeCodeForToken method
func TestExchangeCodeForToken(t *testing.T) {
	// mock token endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}

		if r.Form.Get("code") != "test-code" {
			t.Errorf("expected code=test-code, got %s", r.Form.Get("code"))
		}

		resp := TokenResponse{
			AccessToken: "access-token-123",
			TokenType:   "bearer",
			BlogID:      "123",
			BlogURL:     "https://example.com",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// test exchange code for token
	config := NewOAuthConfig("client-id", "client-secret", "redirect-uri")
	config.TokenEndpoint = server.URL // mock server

	token, err := config.exchangeCodeForToken("test-code")
	if err != nil {
		t.Fatalf("exchangeCodeForToken failed: %v", err)
	}

	if token != "access-token-123" {
		t.Errorf("expected token=access-token-123, got %s", token)
	}
}

// TestCallbackHandler() - test the callbackHandler method
func TestCallbackHandler(t *testing.T) {
	config := NewOAuthConfig("id", "secret", "uri")
	config.state = "valid-state"

	// test success
	req, _ := http.NewRequest("GET", "/callback?code=123&state=valid-state", nil)
	w := httptest.NewRecorder()
	go func() {
		code := <-config.codeChan
		if code != "123" {
			t.Errorf("expected code 123, got %s", code)
		}
	}()

	config.callbackHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// test invalid state
	req, _ = http.NewRequest("GET", "/callback?code=123&state=invalid", nil)
	w = httptest.NewRecorder()
	go func() {
		err := <-config.errChan
		if err == nil {
			t.Error("expected error for invalid state")
		}
	}()

	config.callbackHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	// test error param
	req, _ = http.NewRequest("GET", "/callback?error=access_denied", nil)
	w = httptest.NewRecorder()
	go func() {
		err := <-config.errChan
		if err == nil {
			t.Error("expected error for error param")
		}
	}()

	config.callbackHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
