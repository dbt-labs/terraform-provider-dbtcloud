package validators_test

import (
	"context"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/global_connection/validators"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Minimal schema: only attributes read by BigQueryAuthValidator (see issue #636).
func testBigQueryAuthValidatorSchema() resource_schema.Schema {
	return resource_schema.Schema{
		Attributes: map[string]resource_schema.Attribute{
			"bigquery": resource_schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]resource_schema.Attribute{
					"deployment_env_auth_type":    resource_schema.StringAttribute{Optional: true},
					"application_id":              resource_schema.StringAttribute{Optional: true},
					"application_secret":          resource_schema.StringAttribute{Optional: true},
					"private_key_id":              resource_schema.StringAttribute{Optional: true},
					"private_key":                 resource_schema.StringAttribute{Optional: true},
					"client_email":                resource_schema.StringAttribute{Optional: true},
					"client_id":                   resource_schema.StringAttribute{Optional: true},
					"auth_uri":                    resource_schema.StringAttribute{Optional: true},
					"token_uri":                   resource_schema.StringAttribute{Optional: true},
					"auth_provider_x509_cert_url": resource_schema.StringAttribute{Optional: true},
					"client_x509_cert_url":        resource_schema.StringAttribute{Optional: true},
				},
			},
		},
	}
}

func bigqueryAttributeTypes() map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"deployment_env_auth_type":    tftypes.String,
		"application_id":              tftypes.String,
		"application_secret":          tftypes.String,
		"private_key_id":              tftypes.String,
		"private_key":                 tftypes.String,
		"client_email":                tftypes.String,
		"client_id":                   tftypes.String,
		"auth_uri":                    tftypes.String,
		"token_uri":                   tftypes.String,
		"auth_provider_x509_cert_url": tftypes.String,
		"client_x509_cert_url":        tftypes.String,
	}
}

func strKnown(s string) tftypes.Value {
	return tftypes.NewValue(tftypes.String, s)
}

func strUnknown() tftypes.Value {
	return tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
}

func strNull() tftypes.Value {
	return tftypes.NewValue(tftypes.String, nil)
}

func bigqueryObject(values map[string]tftypes.Value) tftypes.Value {
	return tftypes.NewValue(
		tftypes.Object{AttributeTypes: bigqueryAttributeTypes()},
		values,
	)
}

func configWithBigQuery(bq tftypes.Value) tfsdk.Config {
	s := testBigQueryAuthValidatorSchema()
	root := tftypes.NewValue(
		tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"bigquery": tftypes.Object{AttributeTypes: bigqueryAttributeTypes()},
			},
		},
		map[string]tftypes.Value{"bigquery": bq},
	)
	return tfsdk.Config{Schema: s, Raw: root}
}

func runBigQueryAuthValidator(t *testing.T, cfg tfsdk.Config) *resource.ValidateConfigResponse {
	t.Helper()
	req := resource.ValidateConfigRequest{Config: cfg}
	resp := &resource.ValidateConfigResponse{}
	validators.BigQueryAuthValidator{}.ValidateResource(context.Background(), req, resp)
	return resp
}

// Regression: validate must not treat unknown attribute values (e.g. from for_each over
// variables) as missing — Terraform convention is to defer unknowns to plan/apply.
func TestBigQueryAuthValidator_UnknownServiceAccountFieldsNoError_WIF(t *testing.T) {
	t.Parallel()
	bq := bigqueryObject(map[string]tftypes.Value{
		"deployment_env_auth_type":    strKnown("external-oauth-wif"),
		"application_id":              strKnown("oauth-client-id"),
		"application_secret":          strKnown("oauth-secret"),
		"private_key_id":              strUnknown(),
		"private_key":                 strUnknown(),
		"client_email":                strUnknown(),
		"client_id":                   strUnknown(),
		"auth_uri":                    strUnknown(),
		"token_uri":                   strUnknown(),
		"auth_provider_x509_cert_url": strUnknown(),
		"client_x509_cert_url":        strUnknown(),
	})
	resp := runBigQueryAuthValidator(t, configWithBigQuery(bq))
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no errors for unknown SA fields, got: %s", resp.Diagnostics.Errors())
	}
}

func TestBigQueryAuthValidator_UnknownOAuthFieldsNoError_WIF(t *testing.T) {
	t.Parallel()
	bq := bigqueryObject(map[string]tftypes.Value{
		"deployment_env_auth_type":    strKnown("external-oauth-wif"),
		"application_id":              strUnknown(),
		"application_secret":          strUnknown(),
		"private_key_id":              strKnown("pkid"),
		"private_key":                 strKnown("-----BEGIN KEY-----"),
		"client_email":                strKnown("svc@project.iam.gserviceaccount.com"),
		"client_id":                   strKnown("123"),
		"auth_uri":                    strKnown("https://accounts.google.com/o/oauth2/auth"),
		"token_uri":                   strKnown("https://oauth2.googleapis.com/token"),
		"auth_provider_x509_cert_url": strKnown("https://www.googleapis.com/oauth2/v1/certs"),
		"client_x509_cert_url":        strKnown("https://www.googleapis.com/robot/v1/metadata/x509/x"),
	})
	resp := runBigQueryAuthValidator(t, configWithBigQuery(bq))
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no errors for unknown OAuth fields, got: %s", resp.Diagnostics.Errors())
	}
}

func TestBigQueryAuthValidator_NullClientEmailError_ServiceAccountJSON(t *testing.T) {
	t.Parallel()
	bq := bigqueryObject(map[string]tftypes.Value{
		"deployment_env_auth_type":    strKnown("service-account-json"),
		"application_id":              strNull(),
		"application_secret":          strNull(),
		"private_key_id":              strKnown("pkid"),
		"private_key":                 strKnown("key"),
		"client_email":                strNull(),
		"client_id":                   strKnown("cid"),
		"auth_uri":                    strKnown("https://accounts.google.com/o/oauth2/auth"),
		"token_uri":                   strKnown("https://oauth2.googleapis.com/token"),
		"auth_provider_x509_cert_url": strKnown("https://www.googleapis.com/oauth2/v1/certs"),
		"client_x509_cert_url":        strKnown("https://www.googleapis.com/robot/v1/metadata/x509/x"),
	})
	resp := runBigQueryAuthValidator(t, configWithBigQuery(bq))
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for null client_email")
	}
}

func TestBigQueryAuthValidator_NullApplicationIDError_WIF(t *testing.T) {
	t.Parallel()
	bq := bigqueryObject(map[string]tftypes.Value{
		"deployment_env_auth_type":    strKnown("external-oauth-wif"),
		"application_id":              strNull(),
		"application_secret":          strKnown("secret"),
		"private_key_id":              strKnown("pkid"),
		"private_key":                 strKnown("key"),
		"client_email":                strKnown("svc@project.iam.gserviceaccount.com"),
		"client_id":                   strKnown("123"),
		"auth_uri":                    strKnown("https://accounts.google.com/o/oauth2/auth"),
		"token_uri":                   strKnown("https://oauth2.googleapis.com/token"),
		"auth_provider_x509_cert_url": strKnown("https://www.googleapis.com/oauth2/v1/certs"),
		"client_x509_cert_url":        strKnown("https://www.googleapis.com/robot/v1/metadata/x509/x"),
	})
	resp := runBigQueryAuthValidator(t, configWithBigQuery(bq))
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for null application_id with external-oauth-wif")
	}
}

func TestBigQueryAuthValidator_AllKnownNoError_ServiceAccountJSON(t *testing.T) {
	t.Parallel()
	bq := bigqueryObject(map[string]tftypes.Value{
		"deployment_env_auth_type":    strKnown("service-account-json"),
		"application_id":              strNull(),
		"application_secret":          strNull(),
		"private_key_id":              strKnown("pkid"),
		"private_key":                 strKnown("key"),
		"client_email":                strKnown("svc@project.iam.gserviceaccount.com"),
		"client_id":                   strKnown("123"),
		"auth_uri":                    strKnown("https://accounts.google.com/o/oauth2/auth"),
		"token_uri":                   strKnown("https://oauth2.googleapis.com/token"),
		"auth_provider_x509_cert_url": strKnown("https://www.googleapis.com/oauth2/v1/certs"),
		"client_x509_cert_url":        strKnown("https://www.googleapis.com/robot/v1/metadata/x509/x"),
	})
	resp := runBigQueryAuthValidator(t, configWithBigQuery(bq))
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no errors, got: %s", resp.Diagnostics.Errors())
	}
}
