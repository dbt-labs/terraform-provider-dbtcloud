package bigquery_credential_test

import (
	"context"
	"testing"

	bqcredential "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/bigquery_credential"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func credentialAttributeTypes() map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"id":                                tftypes.String,
		"is_active":                         tftypes.Bool,
		"project_id":                        tftypes.Number,
		"credential_id":                     tftypes.Number,
		"dataset":                           tftypes.String,
		"num_threads":                       tftypes.Number,
		"connection_id":                     tftypes.Number,
		"auth_type":                         tftypes.String,
		"workload_pool_provider_path":       tftypes.String,
		"service_account_impersonation_url": tftypes.String,
	}
}

func credentialConfig(connectionID, authType, workloadPoolProviderPath tftypes.Value) tfsdk.Config {
	raw := tftypes.NewValue(
		tftypes.Object{AttributeTypes: credentialAttributeTypes()},
		map[string]tftypes.Value{
			"id":                                tftypes.NewValue(tftypes.String, nil),
			"is_active":                         tftypes.NewValue(tftypes.Bool, true),
			"project_id":                        tftypes.NewValue(tftypes.Number, 123),
			"credential_id":                     tftypes.NewValue(tftypes.Number, nil),
			"dataset":                           tftypes.NewValue(tftypes.String, "my_dataset"),
			"num_threads":                       tftypes.NewValue(tftypes.Number, 4),
			"connection_id":                     connectionID,
			"auth_type":                         authType,
			"workload_pool_provider_path":       workloadPoolProviderPath,
			"service_account_impersonation_url": tftypes.NewValue(tftypes.String, nil),
		},
	)
	return tfsdk.Config{Schema: bqcredential.BigQueryResourceSchema, Raw: raw}
}

func runCredentialValidateConfig(t *testing.T, cfg tfsdk.Config) *resource.ValidateConfigResponse {
	t.Helper()
	r, ok := bqcredential.BigqueryCredentialResource().(resource.ResourceWithValidateConfig)
	if !ok {
		t.Fatal("bigquery credential resource does not implement ResourceWithValidateConfig")
	}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: cfg}, resp)
	return resp
}

func runCredentialAttributeValidators(
	t *testing.T,
	cfg tfsdk.Config,
	name string,
	value string,
) *validator.StringResponse {
	t.Helper()
	attribute, ok := bqcredential.BigQueryResourceSchema.Attributes[name].(resource_schema.StringAttribute)
	if !ok {
		t.Fatalf("%s is not a string attribute", name)
	}

	req := validator.StringRequest{
		Path:           path.Root(name),
		PathExpression: path.MatchRoot(name),
		Config:         cfg,
		ConfigValue:    types.StringValue(value),
	}
	resp := &validator.StringResponse{}
	for _, v := range attribute.Validators {
		v.ValidateString(context.Background(), req, resp)
	}
	return resp
}

func TestBigQueryCredentialValidateConfig_WIFWithoutWorkloadPoolProviderPath(t *testing.T) {
	t.Parallel()
	resp := runCredentialValidateConfig(t, credentialConfig(
		tftypes.NewValue(tftypes.Number, 456),
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, nil),
	))
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for missing workload_pool_provider_path with external-oauth-wif")
	}
}

func TestBigQueryCredentialValidateConfig_WIFWithWorkloadPoolProviderPath(t *testing.T) {
	t.Parallel()
	resp := runCredentialValidateConfig(t, credentialConfig(
		tftypes.NewValue(tftypes.Number, 456),
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, "projects/1/locations/global/workloadIdentityPools/p/providers/pr"),
	))
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no errors, got: %s", resp.Diagnostics.Errors())
	}
}

func TestBigQueryCredentialValidateConfig_UnknownWorkloadPoolProviderPath(t *testing.T) {
	t.Parallel()
	resp := runCredentialValidateConfig(t, credentialConfig(
		tftypes.NewValue(tftypes.Number, 456),
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	))
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected unknown values to be deferred to apply, got: %s", resp.Diagnostics.Errors())
	}
}

func TestBigQueryCredentialValidateConfig_ServiceAccountJSONWithoutWIFFields(t *testing.T) {
	t.Parallel()
	resp := runCredentialValidateConfig(t, credentialConfig(
		tftypes.NewValue(tftypes.Number, 456),
		tftypes.NewValue(tftypes.String, "service-account-json"),
		tftypes.NewValue(tftypes.String, nil),
	))
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no errors, got: %s", resp.Diagnostics.Errors())
	}
}

func TestBigQueryCredentialWIFAttributes_RequireConnectionID(t *testing.T) {
	t.Parallel()

	withoutConnection := credentialConfig(
		tftypes.NewValue(tftypes.Number, nil),
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, "projects/1/locations/global/workloadIdentityPools/p/providers/pr"),
	)
	withConnection := credentialConfig(
		tftypes.NewValue(tftypes.Number, 456),
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, "projects/1/locations/global/workloadIdentityPools/p/providers/pr"),
	)

	for name, value := range map[string]string{
		"auth_type":                         "external-oauth-wif",
		"workload_pool_provider_path":       "projects/1/locations/global/workloadIdentityPools/p/providers/pr",
		"service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/svc:generateAccessToken",
	} {
		if resp := runCredentialAttributeValidators(t, withoutConnection, name, value); !resp.Diagnostics.HasError() {
			t.Errorf("expected error for %s without connection_id", name)
		}
		if resp := runCredentialAttributeValidators(t, withConnection, name, value); resp.Diagnostics.HasError() {
			t.Errorf("expected no errors for %s with connection_id, got: %s", name, resp.Diagnostics.Errors())
		}
	}
}

// The Semantic Layer credential resource reuses these attributes and its API does not
// accept the v1 WIF fields, so they must not leak into its schema.
func TestBigQueryCredentialSemanticLayerAttributes_ExcludeWIFFields(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"auth_type",
		"workload_pool_provider_path",
		"service_account_impersonation_url",
	} {
		if _, ok := bqcredential.SemanticLayerAttributes[name]; ok {
			t.Errorf("%s should not be part of the Semantic Layer credential attributes", name)
		}
		if _, ok := bqcredential.BigQueryResourceSchema.Attributes[name]; !ok {
			t.Errorf("%s should be part of the BigQuery credential resource schema", name)
		}
	}
}
