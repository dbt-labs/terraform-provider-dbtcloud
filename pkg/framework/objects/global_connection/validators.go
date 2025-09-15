package global_connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// bigqueryTimeoutForV1Validator validates that the timeout is not set for bigquery_v1 adapter.
type bigqueryTimeoutForV1Validator struct{}

func (v *bigqueryTimeoutForV1Validator) Description(ctx context.Context) string {
	return "if adapter_version_override is bigquery_v1, timeout_seconds must not be set > 0"
}

func (v *bigqueryTimeoutForV1Validator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *bigqueryTimeoutForV1Validator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	// If the entire bigquery object is unknown or null, we can't validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var bigquery BigQueryConfig
	diags := req.ConfigValue.As(ctx, &bigquery, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if bigquery.AdapterVersionOverride.IsNull() || bigquery.AdapterVersionOverride.IsUnknown() {
		return
	}

	if bigquery.TimeoutSeconds.IsNull() || bigquery.TimeoutSeconds.IsUnknown() {
		return
	}

	if bigquery.AdapterVersionOverride.ValueString() == "bigquery_v1" && !bigquery.TimeoutSeconds.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path.AtName("timeout_seconds"),
			"Invalid 'timeout_seconds' for adapter 'bigquery_v1'",
			"The `timeout_seconds` field will not be taken into consideration for adapter `bigquery_v1`. Please remove it or set it to 0.",
		)
	}
}

func BigqueryTimeoutForV1() validator.Object {
	return &bigqueryTimeoutForV1Validator{}
}
