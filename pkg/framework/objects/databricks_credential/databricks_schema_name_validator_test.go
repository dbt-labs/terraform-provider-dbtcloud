package databricks_credential

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestSchemaNameValidator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		value       string
		expectError bool
		adapterType string
	}{
		{"valid simple", "my_schema", false, "databricks"},
		{"valid with dot on dbx", "schema.test", true, "databricks"},
		{"valid with space", "my schema", false, "databricks"},
		{"invalid forward slash", "schema/test", true, "databricks"},
		{"invalid backslash", "schema\\test", true, "databricks"},
		{"valid simple", "my_schema", false, "spark"},
		{"valid with dot", "schema.test", false, "spark"},
		{"valid with space", "my schema", false, "spark"},
		{"invalid forward slash", "schema/test", true, "spark"},
		{"invalid backslash", "schema\\test", true, "spark"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock configuration with both schema and adapter_type
			configValue := tftypes.NewValue(
				tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"schema":       tftypes.String,
						"adapter_type": tftypes.String,
					},
				},
				map[string]tftypes.Value{
					"schema":       tftypes.NewValue(tftypes.String, tc.value),
					"adapter_type": tftypes.NewValue(tftypes.String, tc.adapterType),
				},
			)

			config := tfsdk.Config{
				Raw:    configValue,
				Schema: DatabricksResourceSchema, // Use the actual schema from schema.go
			}

			request := validator.StringRequest{
				Path:           path.Root("schema"),
				PathExpression: path.MatchRoot("schema"),
				ConfigValue:    types.StringValue(tc.value),
				Config:         config,
			}
			response := validator.StringResponse{}

			DatabricksSchemaNameValidator{}.ValidateString(context.Background(), request, &response)

			if tc.expectError && !response.Diagnostics.HasError() {
				t.Fatalf("expected error for value %s, but got none", tc.value)
			}

			if !tc.expectError && response.Diagnostics.HasError() {
				t.Fatalf("unexpected error for value %s: %s", tc.value, response.Diagnostics.Errors()[0].Summary())
			}
		})
	}
}
