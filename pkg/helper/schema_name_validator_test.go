package helper

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSchemaNameValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		val         types.String
		expectError bool
	}

	testCases := []testCase{
		{
			name:        "valid schema name - simple",
			val:         types.StringValue("my_schema"),
			expectError: false,
		},
		{
			name:        "valid schema name - with numbers",
			val:         types.StringValue("schema123"),
			expectError: false,
		},
		{
			name:        "valid schema name - mixed case",
			val:         types.StringValue("MySchema"),
			expectError: false,
		},
		{
			name:        "valid schema name - with parentheses",
			val:         types.StringValue("schema(test)"),
			expectError: false,
		},
		{
			name:        "valid schema name - with quotes",
			val:         types.StringValue(`schema"test"`),
			expectError: false,
		},
		{
			name:        "valid schema name - with single quotes",
			val:         types.StringValue("schema'test'"),
			expectError: false,
		},
		{
			name:        "valid schema name - with curly braces",
			val:         types.StringValue("schema{test}"),
			expectError: false,
		},
		{
			name:        "valid schema name - with space",
			val:         types.StringValue("my schema"),
			expectError: false,
		},
		{
			name:        "invalid schema name - with dot",
			val:         types.StringValue("schema.test"),
			expectError: true,
		},
		{
			name:        "invalid schema name - with forward slash",
			val:         types.StringValue("dbt_prod_schema_databricks_user/token"),
			expectError: true,
		},
		{
			name:        "invalid schema name - with backslash",
			val:         types.StringValue("schema\\test"),
			expectError: true,
		},
		{
			name:        "invalid schema name - with at sign",
			val:         types.StringValue("schema@test"),
			expectError: true,
		},
		{
			name:        "invalid schema name - with hash",
			val:         types.StringValue("schema#test"),
			expectError: true,
		},
		{
			name:        "invalid schema name - with dollar sign",
			val:         types.StringValue("schema$test"),
			expectError: true,
		},
		{
			name:        "invalid schema name - with percent",
			val:         types.StringValue("schema%test"),
			expectError: true,
		},
		{
			name:        "invalid schema name - with ampersand",
			val:         types.StringValue("schema&test"),
			expectError: true,
		},
		{
			name:        "null value - no error",
			val:         types.StringNull(),
			expectError: false,
		},
		{
			name:        "unknown value - no error",
			val:         types.StringUnknown(),
			expectError: false,
		},
		{
			name:        "empty string - no error",
			val:         types.StringValue(""),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    tc.val,
			}
			response := validator.StringResponse{}

			SchemaNameValidator().ValidateString(context.Background(), request, &response)

			if tc.expectError && !response.Diagnostics.HasError() {
				t.Fatalf("expected error for value %s, but got none", tc.val.ValueString())
			}

			if !tc.expectError && response.Diagnostics.HasError() {
				t.Fatalf("unexpected error for value %s: %s", tc.val.ValueString(), response.Diagnostics.Errors()[0].Summary())
			}
		})
	}
}
