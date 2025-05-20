package snowflake_credential

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SemanticLayerCredentialValidator struct{}

func (v SemanticLayerCredentialValidator) Description(ctx context.Context) string {
	return "Validates that `user` and `schema` are provided when `semantic_layer_credential` is false."
}

func (v SemanticLayerCredentialValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that `user` and `schema` are provided when `semantic_layer_credential` is false."
}

func (v SemanticLayerCredentialValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {

	// Check if `semantic_layer_credential` is false
	if !req.ConfigValue.ValueBool() || req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		var semantic_layer_enabled types.Bool
		semanticLayerPath := req.Path.ParentPath().AtName("semantic_layer_credential")
		diags := req.Config.GetAttribute(ctx, semanticLayerPath, &semantic_layer_enabled)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if semantic_layer_enabled.IsNull() || !semantic_layer_enabled.ValueBool() {
			// Validate that `user` is provided
			var userValue types.String
			userPath := req.Path.ParentPath().AtName("user")
			diags = req.Config.GetAttribute(ctx, userPath, &userValue)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if !userValue.IsUnknown() {
				if userValue.IsNull() || userValue.ValueString() == "" {
					resp.Diagnostics.AddError(
						"Missing Required Attribute",
						"`user` must be provided when `semantic_layer_credential` is false.",
					)
				}
			}

			// Validate that `schema` is provided
			var schemaValue types.String
			schemaPath := req.Path.ParentPath().AtName("schema")
			diags = req.Config.GetAttribute(ctx, schemaPath, &schemaValue)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if !schemaValue.IsUnknown() {
				if schemaValue.IsNull() || schemaValue.ValueString() == "" {
					resp.Diagnostics.AddError(
						"Missing Required Attribute",
						"`schema` must be provided when `semantic_layer_credential` is false.",
					)
				}
			}
		}
	}
}
