package acctest_helper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// TestSharedClient_RetriesTransientErrors is a regression test: SharedClient
// used to hardcode DisableRetry: true, so every acceptance test's
// exists/destroy verification call (which uses this client) was single-shot
// with zero tolerance for a transient blip - unlike the real provider client,
// which retries. A single reset connection or 502 from the dbt Cloud API was
// therefore enough to fail an otherwise-passing test. This confirms
// SharedClient now retries transient failures the same way the main client
// does, instead of failing on the first attempt.
func TestSharedClient_RetriesTransientErrors(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := atomic.AddInt32(&requestCount, 1)
		if attempt < 3 {
			// Simulate the transient 502s / connection resets seen from the
			// live API.
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("upstream connect error or disconnect/reset before headers"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": {"code": 200, "is_success": true},
			"data": {"id": "123", "name": "test-webhook", "client_url": "https://example.com", "job_ids": []}
		}`))
	}))
	defer server.Close()

	t.Setenv("DBT_CLOUD_ACCOUNT_ID", "1")
	t.Setenv("DBT_CLOUD_TOKEN", "test-token")
	t.Setenv("DBT_CLOUD_HOST_URL", fmt.Sprintf("%s/api", server.URL))

	client, err := SharedClient()
	if err != nil {
		t.Fatalf("SharedClient returned an unexpected error: %v", err)
	}

	webhook, err := client.GetWebhook("123")
	if err != nil {
		t.Fatalf("regression: GetWebhook failed instead of retrying past the transient 502s: %v", err)
	}
	if webhook.WebhookId != "123" {
		t.Fatalf("expected webhook id 123, got %q", webhook.WebhookId)
	}
	if got := atomic.LoadInt32(&requestCount); got != 3 {
		t.Fatalf("expected exactly 3 attempts (2 failures + 1 success), got %d", got)
	}
}
