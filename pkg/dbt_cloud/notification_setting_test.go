package dbt_cloud

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newNotificationSettingTestClient builds a Client pointed at the provided
// httptest.Server with retries disabled so the test does not block on backoff.
func newNotificationSettingTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	t.Setenv("TF_ACC", "1") // skip credential validation in NewClient

	accountID := int64(123)
	token := "test_token"
	maxRetries := 1
	retryInterval := 0
	timeout := 5
	skipValidation := true

	c, err := NewClient(
		&accountID,
		&token,
		&serverURL,
		&maxRetries,
		&retryInterval,
		nil,
		skipValidation,
		&timeout,
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// TestDeleteNotificationSetting_NotFoundIsResourceNotFoundPrefixed verifies the
// contract the framework Delete relies on: when the notification-settings
// DELETE comes back as a 404 (the row is already gone), the client returns an
// error whose message starts with "resource-not-found". The framework resource
// uses strings.HasPrefix(err.Error(), "resource-not-found") to treat the
// destroy as successful, so both the plain and permission-flavored 404
// responses must keep that prefix or the spurious "Error deleting notification
// setting" diagnostic returns and turns CI runs red.
func TestDeleteNotificationSetting_NotFoundIsResourceNotFoundPrefixed(t *testing.T) {
	cases := []struct {
		name        string
		userMessage string
	}{
		{
			name:        "plain 404",
			userMessage: "The requested resource was not found.",
		},
		{
			name:        "404 with permission hint",
			userMessage: "The requested resource was not found. Please check that you have the proper permissions.",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"data":null,"status":{"code":404,"is_success":false,"user_message":"` + tc.userMessage + `","developer_message":""}}`))
			}))
			defer srv.Close()

			c := newNotificationSettingTestClient(t, srv.URL)

			err := c.DeleteNotificationSetting(42)
			if err == nil {
				t.Fatalf("expected an error from a 404 delete, got nil")
			}
			if !strings.HasPrefix(err.Error(), "resource-not-found") {
				t.Fatalf("expected error to start with %q so the framework Delete guard tolerates it, got: %v", "resource-not-found", err)
			}
		})
	}
}

// TestDeleteNotificationSetting_SuccessReturnsNil verifies a normal 2xx delete
// returns no error.
func TestDeleteNotificationSetting_SuccessReturnsNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newNotificationSettingTestClient(t, srv.URL)

	if err := c.DeleteNotificationSetting(42); err != nil {
		t.Fatalf("expected nil error on successful delete, got: %v", err)
	}
}
