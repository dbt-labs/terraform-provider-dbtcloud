package platform_metadata_credentials

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const testURL = "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/sa@p.iam.gserviceaccount.com:generateAccessToken"

var (
	unknownString = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	nullString    = tftypes.NewValue(tftypes.String, nil)
)

func str(value string) tftypes.Value {
	return tftypes.NewValue(tftypes.String, value)
}

// bigQueryObject returns a raw value of the BigQuery schema: every attribute is null except the
// given ones.
func bigQueryObject(t *testing.T, values map[string]tftypes.Value) tftypes.Value {
	t.Helper()
	objectType, ok := BigQueryPlatformMetadataCredentialSchema.Type().TerraformType(context.Background()).(tftypes.Object)
	if !ok {
		t.Fatal("the schema type is not an object")
	}
	attributes := map[string]tftypes.Value{}
	for name, attributeType := range objectType.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, nil)
	}
	for name, value := range values {
		if _, ok := attributes[name]; !ok {
			t.Fatalf("unknown attribute %s", name)
		}
		attributes[name] = value
	}
	return tftypes.NewValue(objectType, attributes)
}

func validServiceAccountValues() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"connection_id":               tftypes.NewValue(tftypes.Number, 42),
		"private_key":                 str("key"),
		"private_key_id":              str("key-id"),
		"client_email":                str("sa@p.iam.gserviceaccount.com"),
		"client_id":                   str("123"),
		"auth_uri":                    str("https://accounts.google.com/o/oauth2/auth"),
		"token_uri":                   str("https://oauth2.googleapis.com/token"),
		"auth_provider_x509_cert_url": str("https://www.googleapis.com/oauth2/v1/certs"),
		"client_x509_cert_url":        str("https://www.googleapis.com/robot/v1/metadata/x509/sa"),
	}
}

func validWIFValues() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"connection_id":               tftypes.NewValue(tftypes.Number, 42),
		"auth_type":                   str(bigQueryAuthTypeWIF),
		"workload_pool_provider_path": str("//iam.googleapis.com/projects/1/locations/global/workloadIdentityPools/p/providers/p"),
	}
}

func with(base map[string]tftypes.Value, changes map[string]tftypes.Value) map[string]tftypes.Value {
	result := map[string]tftypes.Value{}
	for name, value := range base {
		result[name] = value
	}
	for name, value := range changes {
		result[name] = value
	}
	return result
}

func TestBigQueryPMC_ValidateConfig(t *testing.T) {
	saAuth := map[string]tftypes.Value{"auth_type": str(bigQueryAuthTypeServiceAccountJSON)}

	for name, tc := range map[string]struct {
		values map[string]tftypes.Value
		// each expected error must contain all of these strings; nil means no error
		wantErrors [][]string
	}{
		"service account, auth_type null": {
			values: validServiceAccountValues(),
		},
		"service account, explicit auth_type": {
			values: with(validServiceAccountValues(), saAuth),
		},
		"service account with private_key_wo": {
			values: with(validServiceAccountValues(), map[string]tftypes.Value{
				"private_key":            nullString,
				"private_key_wo":         str("wo"),
				"private_key_wo_version": tftypes.NewValue(tftypes.Number, 1),
			}),
		},
		"service account, unknown values are checked at apply": {
			values: with(validServiceAccountValues(), map[string]tftypes.Value{
				"private_key":                 unknownString,
				"client_email":                unknownString,
				"workload_pool_provider_path": unknownString,
			}),
		},
		"auth_type unknown skips every check": {
			values: map[string]tftypes.Value{
				"auth_type":                   unknownString,
				"workload_pool_provider_path": str("path"),
				"private_key":                 str("key"),
			},
		},
		"service account missing a field, auth_type null": {
			values: with(validServiceAccountValues(), map[string]tftypes.Value{"client_email": nullString}),
			wantErrors: [][]string{
				{"client_email is required", "auth_type defaults to service-account-json"},
			},
		},
		"service account missing a field, explicit auth_type": {
			values: with(validServiceAccountValues(), with(saAuth, map[string]tftypes.Value{"token_uri": nullString})),
			wantErrors: [][]string{
				{"token_uri is required when auth_type is service-account-json."},
			},
		},
		"service account without any key": {
			values: with(validServiceAccountValues(), map[string]tftypes.Value{"private_key": nullString}),
			wantErrors: [][]string{
				{"One of private_key or private_key_wo is required"},
			},
		},
		"service account with WIF fields": {
			values: with(validServiceAccountValues(), map[string]tftypes.Value{
				"workload_pool_provider_path":       str("path"),
				"service_account_impersonation_url": str(testURL),
			}),
			wantErrors: [][]string{
				{"workload_pool_provider_path can only be set when auth_type is external-oauth-wif"},
				{"service_account_impersonation_url can only be set when auth_type is external-oauth-wif"},
			},
		},
		"WIF": {
			values: validWIFValues(),
		},
		"WIF with impersonation URL": {
			values: with(validWIFValues(), map[string]tftypes.Value{"service_account_impersonation_url": str(testURL)}),
		},
		"WIF without workload_pool_provider_path": {
			values: with(validWIFValues(), map[string]tftypes.Value{"workload_pool_provider_path": nullString}),
			wantErrors: [][]string{
				{"workload_pool_provider_path is required when auth_type is external-oauth-wif"},
			},
		},
		"WIF with service account fields": {
			values: with(validWIFValues(), map[string]tftypes.Value{
				"private_key_id":         str("key-id"),
				"private_key":            str("key"),
				"private_key_wo":         str("wo"),
				"private_key_wo_version": tftypes.NewValue(tftypes.Number, 1),
			}),
			wantErrors: [][]string{
				{"private_key_id can only be set when auth_type is service-account-json"},
				{"private_key can only be set when auth_type is service-account-json"},
				{"private_key_wo can only be set when auth_type is service-account-json"},
				{"private_key_wo_version can only be set when auth_type is service-account-json"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, ok := BigQueryPlatformMetadataCredentialResource().(resource.ResourceWithValidateConfig)
			if !ok {
				t.Fatal("the resource does not implement ResourceWithValidateConfig")
			}
			resp := &resource.ValidateConfigResponse{}
			r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{
				Config: tfsdk.Config{
					Schema: BigQueryPlatformMetadataCredentialSchema,
					Raw:    bigQueryObject(t, tc.values),
				},
			}, resp)

			errs := resp.Diagnostics.Errors()
			if len(errs) != len(tc.wantErrors) {
				t.Fatalf("expected %d error(s), got %d: %v", len(tc.wantErrors), len(errs), errs)
			}
			for _, want := range tc.wantErrors {
				if !containsError(errs, want) {
					t.Errorf("no error contains %q, got: %v", want, errs)
				}
			}
		})
	}
}

func containsError(errs diag.Diagnostics, parts []string) bool {
	for _, err := range errs {
		found := true
		for _, part := range parts {
			if !strings.Contains(err.Detail(), part) {
				found = false
				break
			}
		}
		if found {
			return true
		}
	}
	return false
}

func TestImpersonationURLPlanModifier(t *testing.T) {
	stateURL := types.StringValue(testURL)

	for name, tc := range map[string]struct {
		configURL       tftypes.Value
		plannedAuthType tftypes.Value
		// nil means no prior state (create)
		stateAuthType *tftypes.Value
		stateURL      types.String
		// the value before the modifier runs: unknown, since the attribute is computed
		planURL types.String
		want    types.String
	}{
		"set in config": {
			configURL:       str(testURL),
			plannedAuthType: str(bigQueryAuthTypeWIF),
			planURL:         stateURL,
			want:            stateURL,
		},
		"unknown in config": {
			configURL:       unknownString,
			plannedAuthType: str(bigQueryAuthTypeWIF),
			planURL:         types.StringUnknown(),
			want:            types.StringUnknown(),
		},
		"planned auth_type unknown": {
			configURL:       nullString,
			plannedAuthType: unknownString,
			stateAuthType:   ptr(str(bigQueryAuthTypeWIF)),
			stateURL:        stateURL,
			planURL:         types.StringUnknown(),
			want:            types.StringUnknown(),
		},
		"create with a service account key": {
			configURL:       nullString,
			plannedAuthType: str(bigQueryAuthTypeServiceAccountJSON),
			planURL:         types.StringUnknown(),
			want:            types.StringNull(),
		},
		"create with WIF": {
			configURL:       nullString,
			plannedAuthType: str(bigQueryAuthTypeWIF),
			planURL:         types.StringUnknown(),
			want:            types.StringUnknown(),
		},
		"switch to a service account key": {
			configURL:       nullString,
			plannedAuthType: str(bigQueryAuthTypeServiceAccountJSON),
			stateAuthType:   ptr(str(bigQueryAuthTypeWIF)),
			stateURL:        stateURL,
			planURL:         types.StringUnknown(),
			want:            types.StringNull(),
		},
		"switch to WIF": {
			configURL:       nullString,
			plannedAuthType: str(bigQueryAuthTypeWIF),
			stateAuthType:   ptr(str(bigQueryAuthTypeServiceAccountJSON)),
			stateURL:        types.StringNull(),
			planURL:         types.StringUnknown(),
			want:            types.StringUnknown(),
		},
		"WIF unchanged, URL in state": {
			configURL:       nullString,
			plannedAuthType: str(bigQueryAuthTypeWIF),
			stateAuthType:   ptr(str(bigQueryAuthTypeWIF)),
			stateURL:        stateURL,
			planURL:         types.StringUnknown(),
			want:            stateURL,
		},
		"WIF unchanged, no URL in state": {
			configURL:       nullString,
			plannedAuthType: str(bigQueryAuthTypeWIF),
			stateAuthType:   ptr(str(bigQueryAuthTypeWIF)),
			stateURL:        types.StringNull(),
			planURL:         types.StringUnknown(),
			want:            types.StringNull(),
		},
		"service account unchanged": {
			configURL:       nullString,
			plannedAuthType: str(bigQueryAuthTypeServiceAccountJSON),
			stateAuthType:   ptr(str(bigQueryAuthTypeServiceAccountJSON)),
			stateURL:        types.StringNull(),
			planURL:         types.StringUnknown(),
			want:            types.StringNull(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			schema := BigQueryPlatformMetadataCredentialSchema

			state := tfsdk.State{Schema: schema, Raw: tftypes.NewValue(schema.Type().TerraformType(ctx), nil)}
			if tc.stateAuthType != nil {
				stateURLRaw, err := tc.stateURL.ToTerraformValue(ctx)
				if err != nil {
					t.Fatal(err)
				}
				state.Raw = bigQueryObject(t, map[string]tftypes.Value{
					"auth_type":                         *tc.stateAuthType,
					"service_account_impersonation_url": stateURLRaw,
				})
			}

			req := planmodifier.StringRequest{
				Path:        path.Root("service_account_impersonation_url"),
				ConfigValue: types.StringNull(),
				Config: tfsdk.Config{Schema: schema, Raw: bigQueryObject(t, map[string]tftypes.Value{
					"service_account_impersonation_url": tc.configURL,
				})},
				Plan: tfsdk.Plan{Schema: schema, Raw: bigQueryObject(t, map[string]tftypes.Value{
					"auth_type": tc.plannedAuthType,
				})},
				PlanValue:  tc.planURL,
				State:      state,
				StateValue: tc.stateURL,
			}
			if tc.configURL.IsKnown() && !tc.configURL.IsNull() {
				req.ConfigValue = types.StringValue(testURL)
			} else if !tc.configURL.IsKnown() {
				req.ConfigValue = types.StringUnknown()
			}

			resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
			impersonationURLPlanModifier{}.PlanModifyString(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected errors: %v", resp.Diagnostics)
			}
			if !resp.PlanValue.Equal(tc.want) {
				t.Errorf("expected %s, got %s", tc.want, resp.PlanValue)
			}
		})
	}
}

func ptr(value tftypes.Value) *tftypes.Value {
	return &value
}

func TestBigQueryPMC_APIErrorHint(t *testing.T) {
	missingFields := errors.New(`resource-not-found: {"data": {"private_key_id": ["field required"], "client_email": ["field required"]}}`)

	for name, tc := range map[string]struct {
		authType string
		err      error
		wantHint bool
	}{
		"WIF rejected as a service account key": {bigQueryAuthTypeWIF, missingFields, true},
		"service account missing fields":        {bigQueryAuthTypeServiceAccountJSON, missingFields, false},
		"WIF, other error":                      {bigQueryAuthTypeWIF, errors.New("unexpected status 500"), false},
	} {
		t.Run(name, func(t *testing.T) {
			var diags diag.Diagnostics
			addAPIError(&diags, "create", tc.authType, tc.err)
			gotHint := strings.Contains(diags.Errors()[0].Detail(), "requires a bigquery_v1 connection")
			if gotHint != tc.wantHint {
				t.Errorf("expected hint %v, got %v: %s", tc.wantHint, gotHint, diags.Errors()[0].Detail())
			}
		})
	}
}
