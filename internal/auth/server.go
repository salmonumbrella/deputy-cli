package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/salmonumbrella/deputy-cli/internal/api"
	"github.com/salmonumbrella/deputy-cli/internal/secrets"
)

var (
	validInstallName = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	validGeos        = map[string]bool{"au": true, "uk": true, "na": true}
)

// clientLimit tracks attempts for a specific client
type clientLimit struct {
	count   int
	resetAt time.Time
}

// rateLimiter tracks attempts per client IP and endpoint
type rateLimiter struct {
	mu          sync.Mutex
	attempts    map[string]*clientLimit
	maxAttempts int
	window      time.Duration
}

func newRateLimiter(maxAttempts int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts:    make(map[string]*clientLimit),
		maxAttempts: maxAttempts,
		window:      window,
	}
}

func (rl *rateLimiter) check(clientIP, endpoint string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := clientIP + ":" + endpoint
	now := time.Now()

	if limit, exists := rl.attempts[key]; exists && now.After(limit.resetAt) {
		delete(rl.attempts, key)
	}

	if rl.attempts[key] == nil {
		rl.attempts[key] = &clientLimit{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return nil
	}

	rl.attempts[key].count++
	if rl.attempts[key].count > rl.maxAttempts {
		return fmt.Errorf("too many attempts, please try again later")
	}
	return nil
}

func (rl *rateLimiter) startCleanup(interval time.Duration, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				rl.cleanup()
			case <-stop:
				return
			}
		}
	}()
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, limit := range rl.attempts {
		if now.After(limit.resetAt) {
			delete(rl.attempts, key)
		}
	}
}

func getClientIP(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// ValidateInstall validates an install name
func ValidateInstall(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("install name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("install name too long (max 64 characters)")
	}
	if !validInstallName.MatchString(name) {
		return fmt.Errorf("install name contains invalid characters")
	}
	return nil
}

// ValidateGeo validates a geographic region
func ValidateGeo(geo string) error {
	if !validGeos[geo] {
		return fmt.Errorf("invalid region: must be au, uk, or na")
	}
	return nil
}

// ValidateToken validates an API token
func ValidateToken(token string) error {
	if len(token) == 0 {
		return fmt.Errorf("API token cannot be empty")
	}
	if len(token) > 512 {
		return fmt.Errorf("API token too long")
	}
	return nil
}

// SetupResult contains the result of a browser-based setup
type SetupResult struct {
	Install string
	Geo     string
	Error   error
}

// SetupServer handles the browser-based authentication flow
type SetupServer struct {
	result        chan SetupResult
	shutdown      chan struct{}
	shutdownOnce  sync.Once
	stopCleanup   chan struct{}
	pendingResult *SetupResult
	pendingMu     sync.Mutex
	csrfToken     string
	store         secrets.Store
	limiter       *rateLimiter
	validateFn    func(ctx context.Context, install, geo, token string) error
}

// NewSetupServer creates a new setup server
func NewSetupServer(store secrets.Store) (*SetupServer, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	stopCleanup := make(chan struct{})
	limiter := newRateLimiter(10, 15*time.Minute)
	limiter.startCleanup(5*time.Minute, stopCleanup)

	return &SetupServer{
		result:      make(chan SetupResult, 1),
		shutdown:    make(chan struct{}),
		stopCleanup: stopCleanup,
		csrfToken:   hex.EncodeToString(tokenBytes),
		store:       store,
		limiter:     limiter,
	}, nil
}

// SetValidator overrides the default credential validation behavior (useful for tests).
func (s *SetupServer) SetValidator(fn func(ctx context.Context, install, geo, token string) error) {
	s.validateFn = fn
}

func (s *SetupServer) validate(ctx context.Context, install, geo, token string) error {
	if s.validateFn != nil {
		return s.validateFn(ctx, install, geo, token)
	}
	return s.validateCredentials(ctx, install, geo, token)
}

// Start starts the setup server and opens the browser
func (s *SetupServer) Start(ctx context.Context) (*SetupResult, error) {
	defer close(s.stopCleanup)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSetup)
	mux.HandleFunc("/validate", s.handleValidate)
	mux.HandleFunc("/submit", s.handleSubmit)
	mux.HandleFunc("/success", s.handleSuccess)
	mux.HandleFunc("/complete", s.handleComplete)

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		_ = server.Serve(listener)
	}()

	go func() {
		if err := openBrowserFunc(baseURL); err != nil {
			slog.Info("failed to open browser, navigate manually", "url", baseURL)
		}
	}()

	fmt.Printf("Opening browser at %s\n", baseURL)
	fmt.Println("Waiting for authentication...")

	select {
	case result := <-s.result:
		_ = server.Shutdown(context.Background())
		return &result, nil
	case <-ctx.Done():
		_ = server.Shutdown(context.Background())
		return nil, ctx.Err()
	case <-s.shutdown:
		_ = server.Shutdown(context.Background())
		s.pendingMu.Lock()
		defer s.pendingMu.Unlock()
		if s.pendingResult != nil {
			return s.pendingResult, nil
		}
		return nil, fmt.Errorf("setup cancelled")
	}
}

func (s *SetupServer) handleSetup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.New("setup").Parse(setupTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	data := map[string]string{
		"CSRFToken": s.csrfToken,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; font-src https://fonts.gstatic.com; connect-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	if err := tmpl.Execute(w, data); err != nil {
		slog.Error("setup template execution failed", "error", err)
	}
}

func (s *SetupServer) validateCredentials(ctx context.Context, install, geo, token string) error {
	if install == "" || geo == "" || token == "" {
		return fmt.Errorf("install name, region, and token are required")
	}

	creds := &secrets.Credentials{
		Token:   token,
		Install: install,
		Geo:     geo,
	}

	client := newAPIClient(creds)
	if _, err := client.Me().Info(ctx); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	return nil
}

// validatedCredentials holds parsed and validated credentials from a request.
type validatedCredentials struct {
	Install string
	Geo     string
	Token   string
}

// validateRequest performs common request validation (method, CSRF, rate limit,
// body parsing, field validation, credential check). Returns nil and writes
// an HTTP error response if validation fails.
func (s *SetupServer) validateRequest(w http.ResponseWriter, r *http.Request, endpoint string) *validatedCredentials {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return nil
	}

	providedToken := r.Header.Get("X-CSRF-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return nil
	}

	clientIP := getClientIP(r)
	if err := s.limiter.check(clientIP, endpoint); err != nil {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return nil
	}

	var req struct {
		Install string `json:"install"`
		Geo     string `json:"geo"`
		Token   string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "Invalid request body",
		})
		return nil
	}

	req.Install = strings.TrimSpace(req.Install)
	req.Geo = strings.TrimSpace(req.Geo)
	req.Token = strings.TrimSpace(req.Token)

	if err := ValidateInstall(req.Install); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return nil
	}
	if err := ValidateGeo(req.Geo); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return nil
	}
	if err := ValidateToken(req.Token); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return nil
	}

	if err := s.validate(r.Context(), req.Install, req.Geo, req.Token); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return nil
	}

	return &validatedCredentials{
		Install: req.Install,
		Geo:     req.Geo,
		Token:   req.Token,
	}
}

func (s *SetupServer) handleValidate(w http.ResponseWriter, r *http.Request) {
	if creds := s.validateRequest(w, r, "/validate"); creds != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"message": "Connection successful!",
		})
	}
}

func (s *SetupServer) handleSubmit(w http.ResponseWriter, r *http.Request) {
	creds := s.validateRequest(w, r, "/submit")
	if creds == nil {
		return
	}

	stored := &secrets.Credentials{
		Token:     creds.Token,
		Install:   creds.Install,
		Geo:       creds.Geo,
		CreatedAt: time.Now(),
	}

	if err := s.store.Set(stored); err != nil {
		slog.Error("failed to save credentials", "error", err)
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Failed to save credentials: %v", err),
		})
		return
	}

	s.pendingMu.Lock()
	s.pendingResult = &SetupResult{
		Install: creds.Install,
		Geo:     creds.Geo,
	}
	s.pendingMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"install": creds.Install,
		"geo":     creds.Geo,
	})
}

func (s *SetupServer) handleSuccess(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("success").Parse(successTemplate)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	s.pendingMu.Lock()
	install := ""
	geo := ""
	if s.pendingResult != nil {
		install = s.pendingResult.Install
		geo = strings.ToUpper(s.pendingResult.Geo)
	}
	s.pendingMu.Unlock()

	data := map[string]string{
		"Install":   install,
		"Geo":       geo,
		"CSRFToken": s.csrfToken,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; font-src https://fonts.gstatic.com; connect-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	if err := tmpl.Execute(w, data); err != nil {
		slog.Error("success template execution failed", "error", err)
	}
}

func (s *SetupServer) handleComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	providedToken := r.Header.Get("X-CSRF-Token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.csrfToken)) != 1 {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	s.shutdownOnce.Do(func() {
		s.pendingMu.Lock()
		if s.pendingResult != nil {
			s.result <- *s.pendingResult
		}
		s.pendingMu.Unlock()
		close(s.shutdown)
	})
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("JSON encoding failed", "error", err)
	}
}

var openBrowserFunc = openBrowser

var goos = runtime.GOOS

var newAPIClient = api.NewClient

var startCommand = func(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

func openBrowser(url string) error {
	switch goos {
	case "darwin":
		return startCommand("open", url)
	case "linux":
		return startCommand("xdg-open", url)
	case "windows":
		return startCommand("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
}
