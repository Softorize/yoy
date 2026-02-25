package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
)

// OAuth client credentials. Can be overridden at build time via ldflags
// or at runtime via environment variables.
var (
	oauthClientID     = ""
	oauthClientSecret = ""
)

const redirectURI = "http://localhost:%d"

// Yahoo OAuth2 endpoints.
var yahooEndpoint = oauth2.Endpoint{
	AuthURL:  "https://api.login.yahoo.com/oauth2/request_auth",
	TokenURL: "https://api.login.yahoo.com/oauth2/get_token",
}

// Yahoo OAuth2 scopes.
const (
	ScopeMailRead  = "mail-r"
	ScopeMailWrite = "mail-w"
)

// resolveCredentials returns the OAuth client ID and secret.
func resolveCredentials() (string, string, error) {
	id := oauthClientID
	secret := oauthClientSecret

	if v := os.Getenv("YOY_CLIENT_ID"); v != "" {
		id = v
	}
	if v := os.Getenv("YOY_CLIENT_SECRET"); v != "" {
		secret = v
	}

	if id == "" || secret == "" {
		return "", "", fmt.Errorf("OAuth credentials not configured.\n\n" +
			"Set them via environment variables:\n" +
			"  export YOY_CLIENT_ID=your-client-id\n" +
			"  export YOY_CLIENT_SECRET=your-client-secret\n\n" +
			"Or build with ldflags:\n" +
			"  go build -ldflags \"-X github.com/Softorize/yoy/internal/auth.oauthClientID=... -X github.com/Softorize/yoy/internal/auth.oauthClientSecret=...\"")
	}

	return id, secret, nil
}

// OAuthConfig returns an oauth2.Config for the given callback port.
func OAuthConfig(port int) (*oauth2.Config, error) {
	id, secret, err := resolveCredentials()
	if err != nil {
		return nil, err
	}
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURL:  fmt.Sprintf(redirectURI, port),
		Scopes:       []string{ScopeMailRead, ScopeMailWrite},
		Endpoint:     yahooEndpoint,
	}, nil
}

// BrowserLogin performs the full browser-based OAuth2 flow and returns a token.
func BrowserLogin(port int) (*oauth2.Token, error) {
	cfg, err := OAuthConfig(port)
	if err != nil {
		return nil, err
	}

	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errCh <- fmt.Errorf("auth error: %s", errMsg)
			http.Error(w, "Authentication failed", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "Authentication successful! You can close this tab.")
		codeCh <- code
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("starting callback server on port %d: %w", port, err)
	}

	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(listener) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	if err := openBrowser(authURL); err != nil {
		return nil, fmt.Errorf("opening browser: %w\nPlease open this URL manually:\n%s", err, authURL)
	}

	select {
	case code := <-codeCh:
		token, err := cfg.Exchange(context.Background(), code)
		if err != nil {
			return nil, fmt.Errorf("exchanging code for token: %w", err)
		}
		return token, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(120 * time.Second):
		return nil, fmt.Errorf("authentication timed out after 120 seconds")
	}
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	return cmd.Start()
}
