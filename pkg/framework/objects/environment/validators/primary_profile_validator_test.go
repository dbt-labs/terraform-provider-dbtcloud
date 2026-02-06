package validators_test

import (
	"context"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/environment/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPrimaryProfileValidator_WarnsWhenConflictingFieldsSet(t *testing.T) {
	t.Parallel()

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"primary_profile_id": schema.Int64Attribute{},
			"connection_id":      schema.Int64Attribute{},
			"credential_id":      schema.Int64Attribute{},
			"extended_attributes_id": schema.Int64Attribute{},
		},
	}

	tests := map[string]struct {
		primaryProfileID int64
		connectionID     int64
		credentialID     int64
		extendedAttrID   int64
		expectWarning    bool
	}{
		"profile with connection_id warns": {
			primaryProfileID: 123,
			connectionID:     456,
			credentialID:     0,
			extendedAttrID:   0,
			expectWarning:    true,
		},
		"profile with credential_id warns": {
			primaryProfileID: 123,
			connectionID:     0,
			credentialID:     789,
			extendedAttrID:   0,
			expectWarning:    true,
		},
		"profile with extended_attributes_id warns": {
			primaryProfileID: 123,
			connectionID:     0,
			credentialID:     0,
			extendedAttrID:   101,
			expectWarning:    true,
		},
		"profile with all conflicting fields warns": {
			primaryProfileID: 123,
			connectionID:     456,
			credentialID:     789,
			extendedAttrID:   101,
			expectWarning:    true,
		},
		"profile alone does not warn": {
			primaryProfileID: 123,
			connectionID:     0,
			credentialID:     0,
			extendedAttrID:   0,
			expectWarning:    false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := tfsdk.Config{
				Schema: testSchema,
				Raw: tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"primary_profile_id":     tftypes.Number,
							"connection_id":          tftypes.Number,
							"credential_id":          tftypes.Number,
							"extended_attributes_id": tftypes.Number,
						},
					},
					map[string]tftypes.Value{
						"primary_profile_id":     tftypes.NewValue(tftypes.Number, tc.primaryProfileID),
						"connection_id":          tftypes.NewValue(tftypes.Number, tc.connectionID),
						"credential_id":          tftypes.NewValue(tftypes.Number, tc.credentialID),
						"extended_attributes_id": tftypes.NewValue(tftypes.Number, tc.extendedAttrID),
					},
				),
			}

			req := validator.Int64Request{
				Path:        path.Root("primary_profile_id"),
				ConfigValue: types.Int64Value(tc.primaryProfileID),
				Config:      config,
			}

			resp := &validator.Int64Response{}

			v := validators.PrimaryProfileValidator{}
			v.ValidateInt64(context.Background(), req, resp)

			hasWarning := resp.Diagnostics.WarningsCount() > 0

			if tc.expectWarning && !hasWarning {
				t.Errorf("expected warning but got none")
			}
			if !tc.expectWarning && hasWarning {
				t.Errorf("expected no warning but got: %v", resp.Diagnostics.Warnings())
			}
		})
	}
}

func TestPrimaryProfileValidator_SkipsWhenNull(t *testing.T) {
	t.Parallel()

	req := validator.Int64Request{
		Path:        path.Root("primary_profile_id"),
		ConfigValue: types.Int64Null(),
	}

	resp := &validator.Int64Response{}

	v := validators.PrimaryProfileValidator{}
	v.ValidateInt64(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors but got: %v", resp.Diagnostics.Errors())
	}
	if resp.Diagnostics.WarningsCount() > 0 {
		t.Errorf("expected no warnings but got: %v", resp.Diagnostics.Warnings())
	}
}

func TestPrimaryProfileValidator_SkipsWhenUnknown(t *testing.T) {
	t.Parallel()

	req := validator.Int64Request{
		Path:        path.Root("primary_profile_id"),
		ConfigValue: types.Int64Unknown(),
	}

	resp := &validator.Int64Response{}

	v := validators.PrimaryProfileValidator{}
	v.ValidateInt64(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors but got: %v", resp.Diagnostics.Errors())
	}
	if resp.Diagnostics.WarningsCount() > 0 {
		t.Errorf("expected no warnings but got: %v", resp.Diagnostics.Warnings())
	}
}
