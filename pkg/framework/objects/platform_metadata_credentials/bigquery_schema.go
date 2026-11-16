package platform_metadata_credentials

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	bigQueryAuthTypeServiceAccountJSON = "service-account-json"
	bigQueryAuthTypeWIF                = "external-oauth-wif"
)

// serviceAccountStringAttribute returns a plain, non-secret field of the service account JSON key.
func serviceAccountStringAttribute(jsonKey string) schema.StringAttribute {
	return schema.StringAttribute{
		Description: "The `" + jsonKey + "` field of the service account JSON key. Required when `auth_type` is `" +
			bigQueryAuthTypeServiceAccountJSON + "`.",
		Optional:   true,
		Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
	}
}

// BigQueryPlatformMetadataCredentialSchema returns the schema for BigQuery platform metadata credentials
var BigQueryPlatformMetadataCredentialSchema = schema.Schema{
	Description: helper.DocString(
		`Manages BigQuery platform metadata credentials for external metadata ingestion in dbt Cloud.

This resource configures credentials that allow dbt Cloud to connect directly to your BigQuery project
to ingest metadata outside of normal dbt project runs. This enables features like:

- **Catalog Ingestion**: Ingest metadata about tables/views not defined in dbt
- **Cost Optimization**: Query warehouse cost and performance data
- **Cost Insights**: Enhanced cost visibility and analysis

Two authentication types are supported:

- ~~~service-account-json~~~ (default): the fields of a service account JSON key. Works with
~~~bigquery_v0~~~ and ~~~bigquery_v1~~~ connections.
- ~~~external-oauth-wif~~~: Workload Identity Federation. Requires a ~~~bigquery_v1~~~ connection
(~~~use_latest_adapter = true~~~ on ~~~dbtcloud_global_connection~~~).

The private key can be rotated and ~~~auth_type~~~ can be changed in place; neither recreates the credential.

~> **Note:** At least one of ~~~catalog_ingestion_enabled~~~, ~~~cost_optimization_enabled~~~, or
~~~cost_insights_enabled~~~ must be enabled for the credential to be usable.

~> **Note:** The ~~~connection_id~~~ cannot be changed after creation. To use a different connection,
you must destroy and recreate the resource.

~> **Note:** The private key is never read back from dbt Cloud. After importing a credential that uses
~~~service-account-json~~~, the first apply sends the private key again. This is expected.`,
	),
	Attributes: mergeAttributes(commonAttributes(), map[string]schema.Attribute{
		"auth_type": schema.StringAttribute{
			Description: "The authentication type: `" + bigQueryAuthTypeServiceAccountJSON + "` (default) or `" +
				bigQueryAuthTypeWIF + "`. `" + bigQueryAuthTypeWIF + "` requires a `bigquery_v1` connection. " +
				"Changing it updates the credential in place.",
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(bigQueryAuthTypeServiceAccountJSON),
			Validators: []validator.String{
				stringvalidator.OneOf(bigQueryAuthTypeServiceAccountJSON, bigQueryAuthTypeWIF),
			},
		},
		"private_key": schema.StringAttribute{
			Description: "The `private_key` field of the service account JSON key. Consider using `private_key_wo` " +
				"instead, which is not stored in state. Exactly one of `private_key` and `private_key_wo` is required " +
				"when `auth_type` is `" + bigQueryAuthTypeServiceAccountJSON + "`.",
			Optional:  true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("private_key_wo")),
				helper.PreferWriteOnlyAttributeValidator{WriteOnlyAttributeName: "private_key_wo"},
			},
		},
		"private_key_wo": schema.StringAttribute{
			Description: "Write-only alternative to `private_key`. The value is not stored in state. " +
				"Requires `private_key_wo_version`.",
			Optional:  true,
			WriteOnly: true,
			Sensitive: true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("private_key_wo_version")),
			},
		},
		"private_key_wo_version": schema.Int64Attribute{
			Description: "Version number for `private_key_wo`. Increment it to send a new `private_key_wo`, and " +
				"whenever `private_key_wo` or `private_key_id` changes. To switch to the plain `private_key`, remove " +
				"both `private_key_wo` and `private_key_wo_version`; the next apply sends the key again.",
			Optional: true,
			Validators: []validator.Int64{
				int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("private_key_wo")),
			},
		},
		"private_key_id":              serviceAccountStringAttribute("private_key_id"),
		"client_email":                serviceAccountStringAttribute("client_email"),
		"client_id":                   serviceAccountStringAttribute("client_id"),
		"auth_uri":                    serviceAccountStringAttribute("auth_uri"),
		"token_uri":                   serviceAccountStringAttribute("token_uri"),
		"auth_provider_x509_cert_url": serviceAccountStringAttribute("auth_provider_x509_cert_url"),
		"client_x509_cert_url":        serviceAccountStringAttribute("client_x509_cert_url"),
		"workload_pool_provider_path": schema.StringAttribute{
			Description: "The full resource path of the workload identity pool provider. Required when " +
				"`auth_type` is `" + bigQueryAuthTypeWIF + "`.",
			Optional:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
		},
		"service_account_impersonation_url": schema.StringAttribute{
			Description: "The service account impersonation URL, used when `auth_type` is `" + bigQueryAuthTypeWIF +
				"`. When it is not set and the connection has `impersonate_service_account`, dbt Cloud fills it in. " +
				"Once set, removing it from the configuration does not clear it.",
			Optional:   true,
			Computed:   true,
			Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			PlanModifiers: []planmodifier.String{
				impersonationURLPlanModifier{},
			},
		},
	}),
}
