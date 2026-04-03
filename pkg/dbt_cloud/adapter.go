package dbt_cloud

type Adapter struct {
	ID                      *int            `json:"id,omitempty"`
	AccountID               int64           `json:"account_id"`
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
