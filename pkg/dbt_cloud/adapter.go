package dbt_cloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Adapter struct {
	ID                      *int            `json:"id,omitempty"`
	AccountID               int             `json:"account_id"`
	ProjectID               int             `json:"project_id"`
	CreatedByID             *int            `json:"created_by_id,omitempty"`
	CreatedByServiceTokenID *int            `json:"created_by_service_token_id,omitempty"`
	Metadata                AdapterMetadata `json:"metadata_json"`
	State                   int             `json:"state"`
	AdapterVersion          string          `json:"adapter_version"`
	CreatedAt               *string         `json:"created_at,omitempty"`
	UpdatedAt               *string         `json:"updated_at,omitempty"`
}

type AdapterMetadata struct {
	Title     string `json:"title"`
	DocsLink  string `json:"docs_link"`
	ImageLink string `json:"image_link"`
}

type AdapterResponse struct {
	Data   Adapter        `json:"data"`
	Status ResponseStatus `json:"status"`
}

type AdapterCredentialFieldMetadataValidation struct {
	Required bool `json:"required"`
}

// Value can actually be a string or an int (for threads)
type AdapterCredentialField struct {
	Metadata AdapterCredentialFieldMetadata `json:"metadata"`
	Value    interface{}                    `json:"value"`
}

type AdapterCredentialDetails struct {
	Fields      map[string]AdapterCredentialField `json:"fields"`
	Field_Order []string                          `json:"field_order"`
}

type AdapterCredentialFieldMetadata struct {
	Label        string                                   `json:"label"`
	Description  string                                   `json:"description"`
	Field_Type   string                                   `json:"field_type"`
	Encrypt      bool                                     `json:"encrypt"`
	Overrideable bool                                     `json:"overrideable"`
	Options      []AdapterCredentialFieldMetadataOptions  `json:"options,omitempty"`
	Validation   AdapterCredentialFieldMetadataValidation `json:"validation"`
}

type AdapterCredentialFieldMetadataOptions struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func createGenericAdapter(c *Client, newAdapter Adapter, projectID int) (*int, error) {
	currentUser, err := c.GetConnectedUser()
	if err != nil {

		// if GetConnectedUser is the following specific error, it means that the user is using a service token
		// as there is no way to get the current token ID, we always use 1
		if strings.Contains(err.Error(), "This endpoint cannot be accessed with a service token") {

			serviceTokenID := 1
			newAdapter.CreatedByServiceTokenID = &serviceTokenID
		} else {
			// if the error is different, return it
			return nil, err
		}
	} else {
		// if there is no error, the user is using a user token
		newAdapter.CreatedByID = &currentUser.ID
	}

	newAdapterData, err := json.Marshal(newAdapter)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"%s/v3/accounts/%s/projects/%s/adapters/",
			c.HostURL,
			strconv.Itoa(c.AccountID),
			strconv.Itoa(projectID),
		),
		strings.NewReader(string(newAdapterData)),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	adapterResponse := AdapterResponse{}
	err = json.Unmarshal(body, &adapterResponse)
	if err != nil {
		return nil, err
	}

	return adapterResponse.Data.ID, nil
}
