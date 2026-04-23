package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

const (
	authURL     = "https://api.prod.whoop.com/oauth/oauth2/auth"
	tokenURL    = "https://api.prod.whoop.com/oauth/oauth2/token"
	redirectURL = "http://localhost:8080/callback"
)

var scopes = []string{
	"read:profile",
	"read:body_measurement",
	"read:cycles",
	"read:recovery",
	"read:sleep",
	"read:workout",
	"offline",
}

type TokenStore struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "whoop-tui")
	return dir, os.MkdirAll(dir, 0700)
}

func tokenPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token.json"), nil
}

func newOAuthConfig(clientID, clientSecret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		RedirectURL: redirectURL,
		Scopes:      scopes,
	}
}

func generateState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func SaveToken(token *oauth2.Token) error {
	path, err := tokenPath()
	if err != nil {
		return err
	}
	store := TokenStore{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadToken() (*oauth2.Token, error) {
	path, err := tokenPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var store TokenStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &oauth2.Token{
		AccessToken:  store.AccessToken,
		RefreshToken: store.RefreshToken,
		TokenType:    store.TokenType,
		Expiry:       store.Expiry,
	}, nil
}

const privacyHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Privacy Policy — WHOOP TUI</title>
<style>body{font-family:system-ui;max-width:700px;margin:40px auto;padding:0 20px;background:#1a1a2e;color:#e0e0e0;line-height:1.6}h1{color:#16c79a}h2{color:#f5c542}code{background:#2a2a3e;padding:2px 6px;border-radius:3px}</style>
</head><body>
<h1>Privacy Policy — WHOOP TUI</h1>
<p><strong>Last updated:</strong> April 23, 2026</p>
<h2>Overview</h2>
<p>WHOOP TUI is a personal, open-source terminal application that displays your WHOOP data locally on your computer. It is not a hosted service.</p>
<h2>Data Collection</h2>
<p>WHOOP TUI does <strong>not</strong> collect, store, transmit, or share any personal data with third parties. All data retrieved from the WHOOP API is displayed locally in your terminal and stored only on your machine.</p>
<h2>Data Storage</h2>
<ul>
<li><strong>OAuth tokens</strong> are stored locally in <code>~/.config/whoop-tui/</code> on your machine.</li>
<li><strong>No data is sent</strong> to any server other than the official WHOOP API (<code>api.prod.whoop.com</code>).</li>
<li><strong>No analytics, telemetry, or tracking</strong> of any kind is used.</li>
</ul>
<h2>WHOOP API Access</h2>
<p>This application uses the WHOOP Developer API with your explicit authorization via OAuth 2.0. You can revoke access at any time through your WHOOP account settings.</p>
<h2>Scopes Requested</h2>
<ul>
<li><code>read:profile</code> — Your name and email</li>
<li><code>read:body_measurement</code> — Height, weight, max heart rate</li>
<li><code>read:cycles</code> — Daily strain and cycle data</li>
<li><code>read:recovery</code> — Recovery scores, HRV, resting heart rate</li>
<li><code>read:sleep</code> — Sleep stages and performance</li>
<li><code>read:workout</code> — Workout activity data</li>
<li><code>offline</code> — Refresh tokens for persistent login</li>
</ul>
<h2>Contact</h2>
<p>If you have questions about this privacy policy, open an issue on the project's GitHub repository.</p>
</body></html>`

func servePrivacyPolicy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, privacyHTML)
}

func ServePrivacyPolicy() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/privacy", servePrivacyPolicy)
	fmt.Println("Privacy policy available at: http://localhost:8080/privacy")
	fmt.Println("Press Ctrl+C to stop.")
	return http.ListenAndServe(":8080", mux)
}

func Authenticate(clientID, clientSecret string) (*oauth2.Token, error) {
	cfg := newOAuthConfig(clientID, clientSecret)

	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("generating state: %w", err)
	}

	tokenCh := make(chan *oauth2.Token, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/privacy", servePrivacyPolicy)
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			errCh <- fmt.Errorf("state mismatch")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "no code", http.StatusBadRequest)
			errCh <- fmt.Errorf("no authorization code received")
			return
		}

		token, err := cfg.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, "token exchange failed", http.StatusInternalServerError)
			errCh <- fmt.Errorf("token exchange: %w", err)
			return
		}

		fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:system-ui;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#1a1a2e;color:#16c79a"><h1>Authenticated! You can close this tab.</h1></body></html>`)
		tokenCh <- token
	})

	server := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("\nOpen this URL in your browser to authenticate:\n\n  %s\n\nWaiting for callback...\n", authURL)

	var token *oauth2.Token
	select {
	case token = <-tokenCh:
	case err := <-errCh:
		server.Shutdown(context.Background())
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("authentication timed out")
	}

	server.Shutdown(context.Background())

	if err := SaveToken(token); err != nil {
		return nil, fmt.Errorf("saving token: %w", err)
	}

	return token, nil
}

func GetClient(clientID, clientSecret string) (*http.Client, error) {
	cfg := newOAuthConfig(clientID, clientSecret)

	token, err := LoadToken()
	if err != nil {
		token, err = Authenticate(clientID, clientSecret)
		if err != nil {
			return nil, err
		}
	}

	src := cfg.TokenSource(context.Background(), token)

	newToken, err := src.Token()
	if err != nil {
		token, err = Authenticate(clientID, clientSecret)
		if err != nil {
			return nil, err
		}
		src = cfg.TokenSource(context.Background(), token)
	} else if newToken.AccessToken != token.AccessToken {
		SaveToken(newToken)
	}

	return oauth2.NewClient(context.Background(), src), nil
}
