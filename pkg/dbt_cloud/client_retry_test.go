package dbt_cloud

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// newRetryTestClient builds a Client pointed at the provided httptest.Server
// with retry timing zeroed out so unit tests do not block on backoff.
func newRetryTestClient(t *testing.T, serverURL string, maxRetries int, retriable []string) *Client {
	t.Helper()
	t.Setenv("TF_ACC", "1") // skip credential validation in NewClient

	accountID := int64(123)
	token := "test_token"
	retryInterval := 0
	timeout := 5
	skipValidation := true

	c, err := NewClient(
		&accountID,
		&token,
		&serverURL,
		&maxRetries,
		&retryInterval,
		retriable,
		skipValidation,
		&timeout,
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// TestDoRequestWithRetry_RetriesTransient5xxAnd429 verifies that the client
// retries each canonical transient status, succeeds on the third attempt, and
// returns the final 2xx body. Without the fix these all returned the first
// 5xx/429 immediately because the per-status early returns short-circuited the
// retry block.
func TestDoRequestWithRetry_RetriesTransient5xxAnd429(t *testing.T) {
	cases := []int{408, 429, 500, 502, 503, 504}
	for _, code := range cases {
		code := code
		t.Run(fmt.Sprintf("status_%d_retries_then_succeeds", code), func(t *testing.T) {
			var attempts atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := attempts.Add(1)
				if n <= 2 {
					http.Error(w, "transient", code)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer srv.Close()

			c := newRetryTestClient(t, srv.URL, 3, nil)
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/test/", nil)
			if err != nil {
				t.Fatalf("NewRequest: %v", err)
			}

			body, err := c.doRequestWithRetry(req)
			if err != nil {
				t.Fatalf("expected success after retries, got: %v", err)
			}
			if string(body) != `{"ok":true}` {
				t.Fatalf("unexpected body: %s", body)
			}
			if got := attempts.Load(); got != 3 {
				t.Fatalf("expected 3 attempts, got %d", got)
			}
		})
	}
}

// TestDoRequestWithRetry_PreservesTerminalErrorOn500Exhaustion locks in the
// back-compat contract for callers that match on the "internal-server-error:"
// prefix. After all retries are exhausted, the surfaced error must still
// contain that substring so existing diagnostic code paths keep working.
func TestDoRequestWithRetry_PreservesTerminalErrorOn500Exhaustion(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`<!doctype html><title>Uh oh!</title>`))
	}))
	defer srv.Close()

	c := newRetryTestClient(t, srv.URL, 2, nil)
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/x/", nil)

	_, err := c.doRequestWithRetry(req)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if !strings.Contains(err.Error(), "internal-server-error:") {
		t.Fatalf("expected wrapped \"internal-server-error:\" prefix for back-compat, got: %v", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("expected 2 attempts, got %d", got)
	}
}

// TestDoRequestWithRetry_PreservesTerminalErrorOn502Exhaustion locks in the
// back-compat contract for the catch-all "unexpected status code" error
// shape, which is what callers see for 502/503/504/etc when retries are
// exhausted.
func TestDoRequestWithRetry_PreservesTerminalErrorOn502Exhaustion(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>502 Bad Gateway</html>"))
	}))
	defer srv.Close()

	c := newRetryTestClient(t, srv.URL, 2, nil)
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/y/", nil)

	_, err := c.doRequestWithRetry(req)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if !strings.Contains(err.Error(), "unexpected status code 502") {
		t.Fatalf("expected wrapped \"unexpected status code 502\" prefix, got: %v", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("expected 2 attempts, got %d", got)
	}
}

// TestDoRequestWithRetry_DoesNotRetry4xx makes sure permission/validation
// errors still surface immediately without burning the retry budget.
func TestDoRequestWithRetry_DoesNotRetry4xx(t *testing.T) {
	cases := []struct {
		name           string
		status         int
		body           string
		errorSubstring string
	}{
		{"401_unauthorized", http.StatusUnauthorized, `{"detail":"bad token"}`, "unauthorized:"},
		{"403_forbidden", http.StatusForbidden, `{"detail":"nope"}`, "forbidden:"},
		{"400_bad_request", http.StatusBadRequest, `{"detail":"bad"}`, "resource-not-found:"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var attempts atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts.Add(1)
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			c := newRetryTestClient(t, srv.URL, 5, nil)
			req, _ := http.NewRequest(http.MethodGet, srv.URL+"/z/", nil)

			_, err := c.doRequestWithRetry(req)
			if err == nil {
				t.Fatal("expected error for 4xx")
			}
			if !strings.Contains(err.Error(), tc.errorSubstring) {
				t.Fatalf("expected error containing %q, got: %v", tc.errorSubstring, err)
			}
			if got := attempts.Load(); got != 1 {
				t.Fatalf("expected exactly 1 attempt for %s, got %d", tc.name, got)
			}
		})
	}
}

// TestDoRequestWithRetry_HonorsConfiguredRetriableStatusCodes verifies that a
// status outside the default retriable set still gets retried when the user
// has explicitly opted in via retriable_status_codes. Picking 418 here so we
// know the behavior comes from the configured list and not the defaults.
func TestDoRequestWithRetry_HonorsConfiguredRetriableStatusCodes(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusTeapot) // 418, not in defaultRetriableHTTPCodes
			_, _ = w.Write([]byte(`steeping`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newRetryTestClient(t, srv.URL, 2, []string{"418"})
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/teapot/", nil)

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		t.Fatalf("expected retry on configured 418, got: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("expected 2 attempts, got %d", got)
	}
}

// TestDoRequestWithRetry_RetriesTransportError ensures that a connection-level
// failure (server closed before response) is treated as a retriable failure.
func TestDoRequestWithRetry_RetriesTransportError(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			// Hijack the connection and close it without writing a response.
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("server does not support hijacking")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatalf("Hijack: %v", err)
			}
			_ = conn.Close()
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newRetryTestClient(t, srv.URL, 3, nil)
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/transport/", nil)

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		t.Fatalf("expected success after transport-error retry, got: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
	if got := attempts.Load(); got < 2 {
		t.Fatalf("expected at least 2 attempts after transport failure, got %d", got)
	}
}

// TestDoRequestWithRetry_DisableRetryStillReturnsSuccess sanity-checks that
// disabling retry produces exactly one attempt and returns the body
// unchanged.
func TestDoRequestWithRetry_DisableRetryStillReturnsSuccess(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newRetryTestClient(t, srv.URL, 3, nil)
	c.DisableRetry = true
	c.MaxRetries = 0 // force the floor-at-1 branch in doRequestWithRetry

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/ok/", nil)
	body, err := c.doRequestWithRetry(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("expected exactly 1 attempt with retry disabled, got %d", got)
	}
}

// TestIsHTTPCodeRetriable covers the helper directly since it now influences
// every API call from the provider.
func TestIsHTTPCodeRetriable(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		configured []string
		want       bool
	}{
		{"transport-error-status-zero", 0, nil, false},
		{"default-500", 500, nil, true},
		{"default-502", 502, nil, true},
		{"default-503", 503, nil, true},
		{"default-504", 504, nil, true},
		{"default-408", 408, nil, true},
		{"default-429", 429, nil, true},
		{"default-501-not-retriable", 501, nil, false},
		{"default-200-not-retriable", 200, nil, false},
		{"default-404-not-retriable", 404, nil, false},
		{"configured-extra-418", 418, []string{"418"}, true},
		{"configured-empty-still-uses-defaults", 502, []string{}, true},
		{"configured-list-without-defaults-still-includes-defaults", 500, []string{"418"}, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := isHTTPCodeRetriable(tc.statusCode, tc.configured)
			if got != tc.want {
				t.Errorf("isHTTPCodeRetriable(%d, %v) = %v, want %v",
					tc.statusCode, tc.configured, got, tc.want)
			}
		})
	}
}

// TestIsErrorRetriable_BackwardsCompatible ensures the legacy helper still
// returns the same answers as the new isHTTPCodeRetriable so any external
// callers that linked against it continue to work.
func TestIsErrorRetriable_BackwardsCompatible(t *testing.T) {
	cases := []struct {
		statusCode int
		configured []string
	}{
		{500, nil},
		{502, nil},
		{418, []string{"418"}},
		{200, nil},
		{404, nil},
	}
	for _, tc := range cases {
		got := isErrorRetriable(tc.statusCode, tc.configured)
		want := isHTTPCodeRetriable(tc.statusCode, tc.configured)
		if got != want {
			t.Errorf("isErrorRetriable(%d, %v) = %v, isHTTPCodeRetriable = %v",
				tc.statusCode, tc.configured, got, want)
		}
	}
}
