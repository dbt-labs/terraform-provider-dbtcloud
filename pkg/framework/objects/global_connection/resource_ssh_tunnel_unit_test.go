package global_connection_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	sshAccountID    = 1
	sshConnectionID = 42
	sshEncryptionID = 7
)

// sshTunnelAPI fakes the connection and encryption endpoints of the API. Like the server since the
// hard delete of encryptions, a connection update that leaves the tunnel disabled removes the
// encryption, so a later call to that encryption is a 404.
type sshTunnelAPI struct {
	mu             sync.Mutex
	adapterVersion string
	tunnelEnabled  bool
	tunnelExists   bool
	// failStatus, when not zero, answers every call to the encryption instead of the real behaviour
	failStatus int
}

func (a *sshTunnelAPI) connection() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	return map[string]interface{}{
		"status": map[string]interface{}{"code": 200, "is_success": true},
		"data": map[string]interface{}{
			"id":                       sshConnectionID,
			"account_id":               sshAccountID,
			"name":                     "test-ssh-tunnel",
			"adapter_version":          a.adapterVersion,
			"is_ssh_tunnel_enabled":    a.tunnelEnabled,
			"private_link_endpoint_id": nil,
			"oauth_configuration_id":   nil,
			"config": map[string]interface{}{
				"hostname": "test.com",
				"port":     1234,
				"dbname":   "my_database",
			},
		},
	}
}

func (a *sshTunnelAPI) encryption(hostname string) map[string]interface{} {
	return map[string]interface{}{
		"id":            sshEncryptionID,
		"account_id":    sshAccountID,
		"connection_id": sshConnectionID,
		"username":      "user",
		"port":          1110,
		"hostname":      hostname,
		"public_key":    "ssh-rsa fake",
		"state":         1,
	}
}

func notFoundResponse() map[string]interface{} {
	return map[string]interface{}{
		"status": map[string]interface{}{
			"code":         404,
			"is_success":   false,
			"user_message": "The requested resource was not found. Please check that you have the proper permissions. If you are seeing this in error, please contact support!",
		},
	}
}

func (a *sshTunnelAPI) server(t *testing.T) *testhelpers.MockServer {
	t.Helper()
	connections := fmt.Sprintf("/v3/accounts/%d/connections/", sshAccountID)
	connection := fmt.Sprintf("%s%d/", connections, sshConnectionID)
	encryptions := fmt.Sprintf("/v2/accounts/%d/encryptions/", sshAccountID)
	encryption := fmt.Sprintf("%s%d/", encryptions, sshEncryptionID)

	hostname := "host2"
	return testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"POST " + connections: func(r *http.Request) (int, interface{}, error) {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			a.mu.Lock()
			a.tunnelEnabled, _ = body["is_ssh_tunnel_enabled"].(bool)
			a.mu.Unlock()
			return http.StatusCreated, a.connection(), nil
		},
		"GET " + connection: func(r *http.Request) (int, interface{}, error) {
			return http.StatusOK, a.connection(), nil
		},
		"PATCH " + connection: func(r *http.Request) (int, interface{}, error) {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			a.mu.Lock()
			if enabled, ok := body["is_ssh_tunnel_enabled"].(bool); ok {
				a.tunnelEnabled = enabled
			}
			if !a.tunnelEnabled {
				a.tunnelExists = false
			}
			a.mu.Unlock()
			return http.StatusOK, a.connection(), nil
		},
		"DELETE " + connection: func(r *http.Request) (int, interface{}, error) {
			return http.StatusOK, map[string]interface{}{
				"status": map[string]interface{}{"code": 200, "is_success": true},
			}, nil
		},
		"POST " + encryptions: func(r *http.Request) (int, interface{}, error) {
			a.mu.Lock()
			a.tunnelEnabled = true
			a.tunnelExists = true
			a.mu.Unlock()
			return http.StatusCreated, map[string]interface{}{"data": a.encryption(hostname)}, nil
		},
		"GET " + encryptions: func(r *http.Request) (int, interface{}, error) {
			a.mu.Lock()
			defer a.mu.Unlock()
			data := []interface{}{}
			if a.tunnelExists {
				data = append(data, a.encryption(hostname))
			}
			return http.StatusOK, map[string]interface{}{"data": data}, nil
		},
		"POST " + encryption: func(r *http.Request) (int, interface{}, error) {
			a.mu.Lock()
			defer a.mu.Unlock()
			if a.failStatus != 0 {
				return a.failStatus, map[string]interface{}{
					"status": map[string]interface{}{"code": a.failStatus, "is_success": false, "user_message": "boom"},
				}, nil
			}
			if !a.tunnelExists {
				return http.StatusNotFound, notFoundResponse(), nil
			}
			a.tunnelExists = false
			return http.StatusOK, map[string]interface{}{"data": a.encryption(hostname)}, nil
		},
	})
}

func sshTunnelConfig(hostURL, adapter string, withTunnel bool) string {
	tunnel := ""
	if withTunnel {
		tunnel = `
    ssh_tunnel = {
      username = "user"
      hostname = "host2"
      port = 1110
    }`
	}
	return fmt.Sprintf(`
provider "dbtcloud" {
  host_url   = "%s"
  token      = "dummy-token"
  account_id = %d
}

resource "dbtcloud_global_connection" "test" {
  name = "test-ssh-tunnel"

  %s = {
    hostname = "test.com"
    port = 1234
    dbname = "my_database"%s
  }
}`, hostURL, sshAccountID, adapter, tunnel)
}

func enableAcceptanceMode(t *testing.T) func() {
	t.Helper()
	original := os.Getenv("TF_ACC")
	os.Setenv("TF_ACC", "1")
	return func() {
		if original == "" {
			os.Unsetenv("TF_ACC")
		} else {
			os.Setenv("TF_ACC", original)
		}
	}
}

// Rule: removing an SSH tunnel relies on the connection update and does not call the tunnel.
// given: a connection with an SSH tunnel, and a server that removes the tunnel when the connection
// update leaves the tunnel disabled (a call to the removed tunnel answers 404)
// when: the ssh_tunnel block is removed from the configuration and applied
// then: the apply succeeds, no call is made to the tunnel, the state has no ssh_tunnel and
// is_ssh_tunnel_enabled is false.
func TestGlobalConnectionSSHTunnel_RemovalDoesNotCallTheTunnel(t *testing.T) {
	defer enableAcceptanceMode(t)()

	for adapter, adapterVersion := range map[string]string{"postgres": "postgres_v0", "redshift": "redshift_v0"} {
		t.Run(adapter, func(t *testing.T) {
			api := &sshTunnelAPI{adapterVersion: adapterVersion}
			srv := api.server(t)
			defer srv.Close()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: sshTunnelConfig(srv.URL, adapter, false)},
					{
						Config: sshTunnelConfig(srv.URL, adapter, true),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("dbtcloud_global_connection.test", "is_ssh_tunnel_enabled", "true"),
							resource.TestCheckResourceAttr("dbtcloud_global_connection.test", adapter+".ssh_tunnel.hostname", "host2"),
						),
					},
					{
						Config: sshTunnelConfig(srv.URL, adapter, false),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("dbtcloud_global_connection.test", "is_ssh_tunnel_enabled", "false"),
							resource.TestCheckNoResourceAttr("dbtcloud_global_connection.test", adapter+".ssh_tunnel.hostname"),
						),
					},
				},
			})

			encryptionPath := fmt.Sprintf("/v2/accounts/%d/encryptions/%d/", sshAccountID, sshEncryptionID)
			if calls := srv.GetCapturedCalls(encryptionPath); len(calls) != 0 {
				t.Errorf("expected no call to the tunnel when it is removed, got %d", len(calls))
			}
		})
	}
}

func (a *sshTunnelAPI) failEncryptionCallsWith(status int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.failStatus = status
}

// Rule: an error on the tunnel call still fails the apply when the tunnel is changed.
// given: a connection with an SSH tunnel and a server that answers the call to the tunnel with a 404
// when: the hostname of the tunnel is changed and applied
// then: the apply fails with the SSH tunnel error, and a later apply succeeds.
func TestGlobalConnectionSSHTunnel_TunnelChangeErrorStillFails(t *testing.T) {
	defer enableAcceptanceMode(t)()

	api := &sshTunnelAPI{adapterVersion: "postgres_v0"}
	srv := api.server(t)
	defer srv.Close()

	changedHostname := strings.Replace(sshTunnelConfig(srv.URL, "postgres", true), `hostname = "host2"`, `hostname = "host3"`, 1)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: sshTunnelConfig(srv.URL, "postgres", false)},
			{Config: sshTunnelConfig(srv.URL, "postgres", true)},
			{
				PreConfig:   func() { api.failEncryptionCallsWith(http.StatusNotFound) },
				Config:      changedHostname,
				ExpectError: regexp.MustCompile(`Error with the SSH Tunnel`),
			},
			{
				// the tunnel calls work again, so the connection can be destroyed afterwards
				PreConfig: func() { api.failEncryptionCallsWith(0) },
				Config:    sshTunnelConfig(srv.URL, "postgres", true),
			},
		},
	})
}
