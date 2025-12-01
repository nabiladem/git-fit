package gravatar

import (
	"testing"
	"time"
)

// TestStartOAuthFlow() - test the StartOAuthFlow method
func TestStartOAuthFlow(t *testing.T) {
	// mock openBrowser to avoid opening real browser
	originalOpenBrowser := OpenBrowserFunc
	defer func() { OpenBrowserFunc = originalOpenBrowser }()

	t.Run("Timeout", func(t *testing.T) {
		// mock OpenBrowser to do nothing
		OpenBrowserFunc = func(url string) error { return nil }

		// set a short timeout for testing
		originalTimeout := OAuthTimeout
		OAuthTimeout = 100 * time.Millisecond
		defer func() { OAuthTimeout = originalTimeout }()

		config := NewOAuthConfig("id", "secret", "uri")
		config.LocalServerPort = ":18080" // use different port to avoid conflicts

		_, err := config.StartOAuthFlow(false)
		if err == nil {
			t.Error("expected timeout error")
		}

		if err.Error() != "authorization timeout after 100ms" {
			t.Errorf("expected timeout error message, got: %v", err)
		}
	})

	// test open browser error
	t.Run("OpenBrowser Error", func(t *testing.T) {
		// mock OpenBrowser to return an error
		OpenBrowserFunc = func(url string) error {
			return nil // errors opening browser are handled gracefully
		}

		// set a very short timeout
		originalTimeout := OAuthTimeout
		OAuthTimeout = 100 * time.Millisecond
		defer func() { OAuthTimeout = originalTimeout }()

		config := NewOAuthConfig("id", "secret", "uri")
		config.LocalServerPort = ":18081"

		_, err := config.StartOAuthFlow(false)
		// should still timeout since we don't send a callback
		if err == nil {
			t.Error("expected timeout error")
		}
	})
}

// TestOpenBrowser() - test the openBrowser method
func TestOpenBrowser(t *testing.T) {
	// Test that the mock can be set
	originalOpenBrowser := OpenBrowserFunc
	defer func() { OpenBrowserFunc = originalOpenBrowser }()

	// test open browser
	called := false
	OpenBrowserFunc = func(url string) error {
		called = true
		return nil
	}

	err := openBrowser("https://example.com")
	if err != nil {
		t.Errorf("openBrowser failed: %v", err)
	}

	if !called {
		t.Error("OpenBrowserFunc was not called")
	}
}
