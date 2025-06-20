package helper

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SemanticLayerCredentialValidator struct {
	FieldName string //the name of the field to check
}

func (v SemanticLayerCredentialValidator) Description(ctx context.Context) string {
	return "Validates that this field is provided when semantic_layer_credential is set to false."
}

func (v SemanticLayerCredentialValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that this field is provided when semantic_layer_credential is set to false."
}

func (v SemanticLayerCredentialValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {

	if req.ConfigValue.IsUnknown() {
		return
	}

	var semantic_layer_enabled types.Bool
	semanticLayerPath := req.Path.ParentPath().AtName("semantic_layer_credential")
	diags := req.Config.GetAttribute(ctx, semanticLayerPath, &semantic_layer_enabled)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if semantic_layer_enabled.IsNull() || !semantic_layer_enabled.ValueBool() {
		if req.ConfigValue.IsNull() || req.ConfigValue.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Missing Required Attribute",
				fmt.Sprintf("`%s` must be provided when `semantic_layer_credential` is false.", v.FieldName),
			)
		}

	}
}
