package dbt_cloud_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud/testutil"
)

// These tests pin down the contract between the dbt Cloud notification-settings
// endpoint and the Terraform provider's recovery behavior. The Read handler
// relies on HandleResourceNotFound recognizing a "resource-not-found"-prefixed
// error to drop a missing resource from state. If the API returns anything else
// — 500, 502, or any other 5xx — the provider must surface it instead of
// silently reconciling, otherwise users get either ghost resources in state or
// silent deletions of resources that still exist server-side.
//
// Background: a customer hit a 502 on DELETE followed by a 500 on the next
// Refresh because the backend was mapping a Django DoesNotExist to a generic
// 500 instead of 404. After the backend fix, GET on a missing setting returns
// 404 and Terraform reconciles cleanly. These tests lock that contract in.

const (
	testAccountID         int64 = 123
	testNotificationID    int64 = 456
	testPrivateAPISegment       = "/private/accounts/123/notification-settings"
)

// notificationSettingMockServer is a configurable test double for the
// /private/accounts/:id/notification-settings endpoint. Each method returns
// the configured status code (and body) on a single matching call. Default
// status is 200 with an empty success body, so unconfigured callers will still
// hit a deterministic path rather than panicking.
type notificationSettingMockServer struct {
	t      *testing.T
	server *httptest.Server

	getStatus    int
	getBody      string
	deleteStatus int
	deleteBody   string
	postStatus   int
	postBody     string
	patchStatus  int
	patchBody    string

	lastRequestMethod string
	lastRequestPath   string
	lastRequestBody   []byte
}

func newNotificationSettingMockServer(t *testing.T) *notificationSettingMockServer {
	m := &notificationSettingMockServer{
		t:            t,
		getStatus:    http.StatusOK,
		deleteStatus: http.StatusNoContent,
		postStatus:   http.StatusCreated,
		patchStatus:  http.StatusOK,
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.serve))
	return m
}

func (m *notificationSettingMockServer) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	m.lastRequestMethod = r.Method
	m.lastRequestPath = r.URL.Path
	m.lastRequestBody = body

	if !strings.HasPrefix(r.URL.Path, testPrivateAPISegment) {
		http.Error(w, "unexpected path "+r.URL.Path, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(m.getStatus)
		_, _ = w.Write([]byte(m.getBody))
	case http.MethodDelete:
		w.WriteHeader(m.deleteStatus)
		_, _ = w.Write([]byte(m.deleteBody))
	case http.MethodPost:
		w.WriteHeader(m.postStatus)
		_, _ = w.Write([]byte(m.postBody))
	case http.MethodPatch:
		w.WriteHeader(m.patchStatus)
		_, _ = w.Write([]byte(m.patchBody))
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (m *notificationSettingMockServer) close() { m.server.Close() }

func (m *notificationSettingMockServer) client() *dbt_cloud.Client {
	return testutil.CreateTestClient(m.server.URL, testAccountID)
}

// not-found-style errors are the only ones HandleResourceNotFound recognizes.
// All other errors must propagate to Terraform.
func errorIsNotFound(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "resource-not-found")
}

func TestGetNotificationSetting_Success(t *testing.T) {
	srv := newNotificationSettingMockServer(t)
	defer srv.close()

	srv.getStatus = http.StatusOK
	srv.getBody = `{
		"status": {"code": 200, "is_success": true},
		"data": {
			"id": 456,
			"account_id": 123,
			"name": "weekly digest",
			"description": "send teams pings",
			"is_active": true,
			"channels": [{"id": 1, "channel_type": "teams"}],
			"rules": [{"id": 2, "trigger_on": "run_errored"}]
		}
	}`

	got, err := srv.client().GetNotificationSetting(testNotificationID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID == nil || *got.ID != testNotificationID {
		t.Fatalf("expected ID %d, got %+v", testNotificationID, got)
	}
	if got.Name != "weekly digest" {
		t.Errorf("expected name %q, got %q", "weekly digest", got.Name)
	}
	if len(got.Channels) != 1 || len(got.Rules) != 1 {
		t.Errorf("expected one channel and one rule, got %d channels and %d rules", len(got.Channels), len(got.Rules))
	}
}

// Regression test: post-backend-fix, a missing setting returns 404 and the
// provider must surface that as a "resource-not-found" error so the Read
// handler removes the resource from state via HandleResourceNotFound.
func TestGetNotificationSetting_404_IsRecognizedAsResourceNotFound(t *testing.T) {
	srv := newNotificationSettingMockServer(t)
	defer srv.close()

	srv.getStatus = http.StatusNotFound
	srv.getBody = `{
		"status": {
			"code": 404,
			"is_success": false,
			"user_message": "Notification setting not found"
		},
		"data": null
	}`

	_, err := srv.client().GetNotificationSetting(testNotificationID)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errorIsNotFound(err) {
		t.Fatalf("expected resource-not-found error, got %q", err.Error())
	}
}

// Regression test: before the backend fix, a missing setting surfaced as a
// generic 500. The provider must NOT silently reconcile that — it should
// propagate the error so Terraform retries or the user investigates, rather
// than dropping a resource from state that may still exist server-side.
//
// This is the exact path that stranded Aramark: Refresh hit a 500 and the
// provider correctly refused to drop the resource. The test pins down that
// behavior so a future "just treat any 5xx as gone" change can't regress it.
func TestGetNotificationSetting_500_IsNotMaskedAsNotFound(t *testing.T) {
	srv := newNotificationSettingMockServer(t)
	defer srv.close()

	srv.getStatus = http.StatusInternalServerError
	srv.getBody = `{
		"status": {
			"code": 500,
			"is_success": false,
			"user_message": "Internal server error getting notification setting"
		}
	}`

	_, err := srv.client().GetNotificationSetting(testNotificationID)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if errorIsNotFound(err) {
		t.Fatalf("500 must not be classified as resource-not-found, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "internal-server-error") {
		t.Errorf("expected internal-server-error tag in error, got %q", err.Error())
	}
}

// Regression test: matches the customer-reported failure mode — DELETE
// returned 502 from the gateway. The next Refresh re-fetched via GET, which
// the gateway also dropped with a 502 protocol error. That must propagate
// (not be silently reconciled) so Terraform retries.
func TestGetNotificationSetting_502_IsNotMaskedAsNotFound(t *testing.T) {
	srv := newNotificationSettingMockServer(t)
	defer srv.close()

	srv.getStatus = http.StatusBadGateway
	srv.getBody = "upstream connect error or disconnect/reset before headers"

	_, err := srv.client().GetNotificationSetting(testNotificationID)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if errorIsNotFound(err) {
		t.Fatalf("502 must not be classified as resource-not-found, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("expected the 502 status to appear in the error, got %q", err.Error())
	}
}

func TestDeleteNotificationSetting_Success(t *testing.T) {
	srv := newNotificationSettingMockServer(t)
	defer srv.close()

	srv.deleteStatus = http.StatusNoContent

	if err := srv.client().DeleteNotificationSetting(testNotificationID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv.lastRequestMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", srv.lastRequestMethod)
	}
	wantPath := fmt.Sprintf("%s/%d/", testPrivateAPISegment, testNotificationID)
	if srv.lastRequestPath != wantPath {
		t.Errorf("expected path %q, got %q", wantPath, srv.lastRequestPath)
	}
}

// Regression test: after the backend fix, DELETE on a missing setting returns
// 404 (mapped from Django DoesNotExist). The provider currently surfaces this
// as an error from Delete — that's fine, since Terraform's destroy plan would
// not have been re-run if the resource was already gone. What matters is that
// it doesn't crash and the error string is recognizable.
func TestDeleteNotificationSetting_404(t *testing.T) {
	srv := newNotificationSettingMockServer(t)
	defer srv.close()

	srv.deleteStatus = http.StatusNotFound
	srv.deleteBody = `{
		"status": {"code": 404, "is_success": false, "user_message": "Notification setting not found"},
		"data": null
	}`

	err := srv.client().DeleteNotificationSetting(testNotificationID)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errorIsNotFound(err) {
		t.Fatalf("expected resource-not-found error, got %q", err.Error())
	}
}

// Regression test for Aramark's original symptom: DELETE returns 502 from the
// gateway. The provider must surface the 502 so the user knows the destroy
// did not complete successfully — silently swallowing it would either leave
// orphaned server-side state or cause Terraform to think the resource is
// gone when it isn't.
func TestDeleteNotificationSetting_502_Surfaces(t *testing.T) {
	srv := newNotificationSettingMockServer(t)
	defer srv.close()

	srv.deleteStatus = http.StatusBadGateway
	srv.deleteBody = "upstream connect error or disconnect/reset before headers"

	err := srv.client().DeleteNotificationSetting(testNotificationID)
	if err == nil {
		t.Fatal("expected a 502 error, got nil")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("expected 502 in error, got %q", err.Error())
	}
}

func TestCreateNotificationSetting_Success(t *testing.T) {
	srv := newNotificationSettingMockServer(t)
	defer srv.close()

	srv.postStatus = http.StatusCreated
	srv.postBody = `{
		"status": {"code": 201, "is_success": true},
		"data": {
			"id": 456,
			"name": "alerts",
			"is_active": true,
			"channels": [],
			"rules": []
		}
	}`

	desc := "first"
	created, err := srv.client().CreateNotificationSetting(dbt_cloud.NotificationSetting{
		Name:        "alerts",
		Description: &desc,
		IsActive:    true,
		AccountID:   999, // The client should strip this; the URL path carries the account.
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created == nil || created.ID == nil || *created.ID != testNotificationID {
		t.Fatalf("expected ID %d, got %+v", testNotificationID, created)
	}

	var sent dbt_cloud.NotificationSetting
	if err := json.Unmarshal(srv.lastRequestBody, &sent); err != nil {
		t.Fatalf("could not unmarshal recorded request body: %v", err)
	}
	if sent.AccountID != 0 {
		t.Errorf("AccountID must be stripped from request body (URL carries it), got %d", sent.AccountID)
	}
	if sent.Name != "alerts" {
		t.Errorf("expected name %q, got %q", "alerts", sent.Name)
	}
}

func TestUpdateNotificationSetting_Success(t *testing.T) {
	srv := newNotificationSettingMockServer(t)
	defer srv.close()

	srv.patchStatus = http.StatusOK
	srv.patchBody = `{
		"status": {"code": 200, "is_success": true},
		"data": {
			"id": 456,
			"name": "renamed",
			"is_active": false,
			"channels": [],
			"rules": []
		}
	}`

	updated, err := srv.client().UpdateNotificationSetting(testNotificationID, dbt_cloud.NotificationSetting{
		Name:     "renamed",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil || updated.Name != "renamed" {
		t.Fatalf("expected updated name %q, got %+v", "renamed", updated)
	}
	if srv.lastRequestMethod != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", srv.lastRequestMethod)
	}
}
