package dbt_cloud

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
