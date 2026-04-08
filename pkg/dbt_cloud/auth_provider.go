package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// AuthProvider represents a dbt Cloud SSO auth provider.
// The API is polymorphic: SAML, Azure AD, and Google Workspace providers
// share this struct, with only the relevant type-specific fields populated.
type AuthProvider struct {
	ID                    *int64  `json:"id,omitempty"`
	AccountID             int64   `json:"account_id"`
	Type                  string  `json:"type"`
	Slug                  string  `json:"slug,omitempty"`
	State                 int     `json:"state,omitempty"`
	AllowPasswordBackdoor bool    `json:"allow_password_backdoor"`
	LoginURL              string  `json:"login_url,omitempty"`
	CreatedAt             string  `json:"created_at,omitempty"`
	UpdatedAt             string  `json:"updated_at,omitempty"`
	CertExpiryDate        *string `json:"cert_expiry_date,omitempty"`

	// SAML / Okta
	EntityID     *string `json:"entity_id,omitempty"`
	SsoURL       *string `json:"sso_url,omitempty"`
	Cert         *string `json:"cert,omitempty"`
	SignRequest  *bool   `json:"sign_request,omitempty"`
	AttributeMap *string `json:"attribute_map,omitempty"`

	// Azure AD (single_tenant, multi_tenant, active_directory) and Google Workspace
	ClientID     *string `json:"client_id,omitempty"`
	ClientSecret *string `json:"client_secret,omitempty"`

	// Azure AD only
	TenantID              *string `json:"tenant_id,omitempty"`
	Domain                *string `json:"domain,omitempty"`
	IncludeIndirectGroups *bool   `json:"include_indirect_groups,omitempty"`
	MaxGroupsToRetrieve   *int    `json:"max_groups_to_retrieve,omitempty"`

	// Google Workspace only
	GsuiteAdminID     *string `json:"gsuite_admin_id,omitempty"`
	AuthorizationURL  *string `json:"authorization_url,omitempty"`
	AdminRefreshToken *string `json:"admin_refresh_token,omitempty"`
}

type AuthProviderResponse struct {
	Data   AuthProvider   `json:"data"`
	Status ResponseStatus `json:"status"`
}

func (c *Client) GetAuthProvider(authProviderID int64) (*AuthProvider, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/v3/accounts/%s/auth-provider/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			authProviderID,
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	resp := AuthProviderResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) CreateAuthProvider(authProvider AuthProvider) (*AuthProvider, error) {
	authProvider.AccountID = c.AccountID

	payload, err := json.Marshal(authProvider)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/auth-provider/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
		),
		strings.NewReader(string(payload)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	resp := AuthProviderResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) UpdateAuthProvider(authProviderID int64, authProvider AuthProvider) (*AuthProvider, error) {
	authProvider.AccountID = c.AccountID

	payload, err := json.Marshal(authProvider)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/auth-provider/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			authProviderID,
		),
		strings.NewReader(string(payload)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}

	resp := AuthProviderResponse{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) DeleteAuthProvider(authProviderID int64) error {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf(
			"%s/v3/accounts/%s/auth-provider/%d/",
			c.HostURL,
			strconv.FormatInt(c.AccountID, 10),
			authProviderID,
		),
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.doRequestWithRetry(req)
	return err
}
