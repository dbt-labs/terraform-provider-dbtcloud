package analytics_credential

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	blockSnowflake  = "snowflake"
	blockDatabricks = "databricks"

	authTypePassword = "password"
	authTypeKeypair  = "keypair"
)

// supportedAdapterBlocks lists the adapter blocks of the schema; exactly one must be set
var supportedAdapterBlocks = []string{blockSnowflake, blockDatabricks}

// adapterVersionToBlock maps the adapter version of a connection to its adapter block
var adapterVersionToBlock = map[string]string{
	"snowflake_v0":  "snowflake",
	"databricks_v0": "databricks",
	"redshift_v0":   "redshift",
	"bigquery_v0":   "bigquery",
	"bigquery_v1":   "bigquery",
}

const resourceDescription = "Analytics credential of a global connection: the warehouse credentials dbt platform uses to read the data of an analytics connection.\n\n" +
	"~> The analytics credentials endpoints are not public yet. They need the `cc-4192-enable-analytics-credentials` flag enabled on the account.\n\n" +
	"The adapter is fixed by the connection: set exactly one adapter block, and it must match the adapter of the connection. Changing the adapter block, or the `auth_type` of an adapter block, recreates the credential. " +
	"Optional attributes of an adapter block cannot be removed once set; recreate the credential to clear them.\n\n" +
	"Secrets are write-only (`*_wo`, with a `*_wo_version` to trigger a rotation) and are never stored in the state. This needs Terraform >= 1.11.\n\n" +
	"Snowflake password authentication is deprecated on the server."

func writeOnlySecret(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:    true,
		WriteOnly:   true,
		Sensitive:   true,
		Description: description + " Write-only: it is never stored in the state. Requires the matching `_wo_version`.",
	}
}

func secretVersion(secret string) schema.Int64Attribute {
	return schema.Int64Attribute{
		Optional:    true,
		Description: "Version of `" + secret + "`. Change it to send a new value of `" + secret + "`; the value alone is not tracked because it is write-only.",
	}
}

func analyticsCredentialSchema() schema.Schema {
	return schema.Schema{
		Description: resourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the analytics credential.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"connection_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the global connection the credential belongs to. Changing it recreates the credential.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"adapter_version": schema.StringAttribute{
				Computed:    true,
				Description: "The adapter version of the credential, taken from the connection.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			blockSnowflake: schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Snowflake credential. Password authentication is deprecated on the server.",
				Attributes: map[string]schema.Attribute{
					"auth_type": schema.StringAttribute{
						Required:    true,
						Description: "The authentication type: `password` (deprecated) or `keypair`. Changing it recreates the credential.",
						Validators: []validator.String{
							stringvalidator.OneOf(authTypePassword, authTypeKeypair),
						},
					},
					"user": schema.StringAttribute{
						Required:    true,
						Description: "The Snowflake user name.",
					},
					"role": schema.StringAttribute{
						Optional:    true,
						Description: "The Snowflake role. It cannot be removed once set.",
					},
					"warehouse": schema.StringAttribute{
						Required:    true,
						Description: "The Snowflake warehouse.",
					},
					"password_wo":                       writeOnlySecret("The password. Required when `auth_type` is `password`, not allowed with `keypair`."),
					"password_wo_version":               secretVersion("password_wo"),
					"private_key_wo":                    writeOnlySecret("The private key. Required when `auth_type` is `keypair`, not allowed with `password`."),
					"private_key_wo_version":            secretVersion("private_key_wo"),
					"private_key_passphrase_wo":         writeOnlySecret("The passphrase of the private key. Only with `keypair`."),
					"private_key_passphrase_wo_version": secretVersion("private_key_passphrase_wo"),
				},
			},
			blockDatabricks: schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Databricks credential.",
				Attributes: map[string]schema.Attribute{
					"token_wo": schema.StringAttribute{
						Required:    true,
						WriteOnly:   true,
						Sensitive:   true,
						Description: "The personal access token. Write-only: it is never stored in the state.",
					},
					"token_wo_version": schema.Int64Attribute{
						Required:    true,
						Description: "Version of `token_wo`. Change it to send a new token; the value alone is not tracked because it is write-only.",
					},
					"catalog": schema.StringAttribute{
						Optional:    true,
						Description: "The Unity Catalog name. It cannot be removed once set.",
					},
				},
			},
		},
	}
}
