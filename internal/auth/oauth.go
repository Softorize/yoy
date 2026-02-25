package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
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

const redirectURI = "https://localhost:%d"

// Yahoo OAuth2 endpoints.
var yahooEndpoint = oauth2.Endpoint{
	AuthURL:  "https://api.login.yahoo.com/oauth2/request_auth",
	TokenURL: "https://api.login.yahoo.com/oauth2/get_token",
}

// Yahoo OAuth2 scopes.
// Yahoo no longer grants mail-r/mail-w to new apps, but IMAP XOAUTH2
// works with a valid token regardless of scope. We use OpenID Connect
// scopes to obtain the token.
const (
	ScopeOpenID = "openid"
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
		Scopes:       []string{ScopeOpenID},
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

	tlsCert, err := selfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("generating TLS certificate: %w", err)
	}

	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})
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

func selfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(1 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return tls.X509KeyPair(certPEM, keyPEM)
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
