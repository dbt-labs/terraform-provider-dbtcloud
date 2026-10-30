package analytics_credential

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	blockSnowflake  = "snowflake"
	blockDatabricks = "databricks"
	blockRedshift   = "redshift"
	blockBigQuery   = "bigquery"

	authTypeServiceAccount = "service-account-json"
	authTypeWIF            = "external-oauth-wif"

	authTypePassword = "password"
	authTypeKeypair  = "keypair"
)

// supportedAdapterBlocks lists the adapter blocks of the schema; exactly one must be set
var supportedAdapterBlocks = []string{blockSnowflake, blockDatabricks, blockRedshift, blockBigQuery}

// adapterVersionToBlock maps the adapter version of a connection to its adapter block
var adapterVersionToBlock = map[string]string{
	"snowflake_v0":  "snowflake",
	"databricks_v0": "databricks",
	"redshift_v0":   "redshift",
	"bigquery_v0":   "bigquery",
	"bigquery_v1":   "bigquery",
}

const resourceDescription = "Analytics credential of a global connection: the shared warehouse credential that read-only users' queries run with when they have no personal credentials. It is shared, so give it the least privilege needed.\n\n" +
	"The adapter is fixed by the connection: set exactly one adapter block, and it must match the adapter of the connection. Changing the adapter block, or the `auth_type` of an adapter block, recreates the credential. " +
	"Optional attributes of an adapter block cannot be removed once set; recreate the credential to clear them.\n\n" +
	"A connection can have one analytics credential, so a replacement on the same connection cannot use `create_before_destroy`. Deleting the credential also clears the analytics connection of the projects that use the connection.\n\n" +
	"Each secret is given either as a write-only value (`*_wo`, with a `*_wo_version` to trigger a rotation), which is never stored in the state and needs Terraform >= 1.11, " +
	"or as a plain value, which is marked sensitive but stored in the state. Prefer the write-only value."

func writeOnlySecret(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:    true,
		WriteOnly:   true,
		Sensitive:   true,
		Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
		Description: description + " Write-only: it is never stored in the state. Requires the matching `_wo_version`.",
	}
}

// plainSecret is the plain alternative of the write-only secret `writeOnly`: it is sensitive but stored in the state
func plainSecret(description, writeOnly string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:  true,
		Sensitive: true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
			stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(writeOnly)),
			helper.PreferWriteOnlyAttributeValidator{WriteOnlyAttributeName: writeOnly},
		},
		Description: description + " Stored in the state, marked sensitive. Prefer `" + writeOnly + "`, which is never stored.",
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
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The Snowflake user name.",
					},
					"role": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The Snowflake role. It cannot be removed once set.",
					},
					"warehouse": schema.StringAttribute{
						Required:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The Snowflake warehouse.",
					},
					"password":                          plainSecret("The password. Required when `auth_type` is `password`, not allowed with `keypair`.", "password_wo"),
					"private_key":                       plainSecret("The private key. Required when `auth_type` is `keypair`, not allowed with `password`.", "private_key_wo"),
					"private_key_passphrase":            plainSecret("The passphrase of the private key. Only with `keypair`.", "private_key_passphrase_wo"),
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
					"token": func() schema.StringAttribute {
						attribute := plainSecret("The personal access token. Required unless `token_wo` is set.", "token_wo")
						attribute.Validators = append(attribute.Validators,
							stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("token_wo")))
						return attribute
					}(),
					"token_wo": schema.StringAttribute{
						Optional:  true,
						WriteOnly: true,
						Sensitive: true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("token_wo_version")),
						},
						Description: "The personal access token, write-only: it is never stored in the state. Required unless `token` is set, and it requires `token_wo_version`.",
					},
					"token_wo_version": func() schema.Int64Attribute {
						attribute := secretVersion("token_wo")
						attribute.Validators = []validator.Int64{
							int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("token_wo")),
						}
						return attribute
					}(),
					"catalog": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The Unity Catalog name. It cannot be removed once set.",
					},
				},
			},
			blockRedshift: schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Redshift credential.",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Required:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The username to connect to the cluster.",
					},
					"password": func() schema.StringAttribute {
						attribute := plainSecret("The password to connect to the cluster. Required unless `password_wo` is set.", "password_wo")
						attribute.Validators = append(attribute.Validators,
							stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("password_wo")))
						return attribute
					}(),
					"password_wo": schema.StringAttribute{
						Optional:  true,
						WriteOnly: true,
						Sensitive: true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo_version")),
						},
						Description: "The password to connect to the cluster, write-only: it is never stored in the state. Required unless `password` is set, and it requires `password_wo_version`.",
					},
					"password_wo_version": func() schema.Int64Attribute {
						attribute := secretVersion("password_wo")
						attribute.Validators = []validator.Int64{
							int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo")),
						}
						return attribute
					}(),
				},
			},
			blockBigQuery: schema.SingleNestedAttribute{
				Optional:    true,
				Description: "BigQuery credential, either a service account (the default) or workload identity federation. A `bigquery_v0` connection only supports service accounts and takes no `auth_type`.",
				Attributes: map[string]schema.Attribute{
					"auth_type": schema.StringAttribute{
						Optional:    true,
						Description: "The authentication type: `service-account-json` (the default on the latest adapter) or `external-oauth-wif`. Not allowed on a `bigquery_v0` connection. Changing it recreates the credential.",
						Validators: []validator.String{
							stringvalidator.OneOf(authTypeServiceAccount, authTypeWIF),
						},
					},
					"execution_project": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The project billed for query execution. It cannot be removed once set.",
					},
					"private_key_id": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The private key ID of the service account. Service account only.",
					},
					"private_key":            plainSecret("The private key of the service account. Required for a service account, not allowed with workload identity federation.", "private_key_wo"),
					"private_key_wo":         writeOnlySecret("The private key of the service account. Required for a service account, not allowed with workload identity federation."),
					"private_key_wo_version": secretVersion("private_key_wo"),
					"client_email": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The client email of the service account. Service account only.",
					},
					"client_id": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The client ID of the service account. Service account only.",
					},
					"auth_uri": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The auth URI of the service account. Service account only.",
					},
					"token_uri": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The token URI of the service account. Service account only.",
					},
					"auth_provider_x509_cert_url": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The auth provider X509 cert URL of the service account. Service account only.",
					},
					"client_x509_cert_url": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The client X509 cert URL of the service account. Service account only.",
					},
					"workload_pool_provider_path": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The audience of the Security Token Service, usually the full resource name of the workload pool provider. Required for `external-oauth-wif`, not allowed otherwise.",
					},
					"service_account_impersonation_url": schema.StringAttribute{
						Optional:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						Description: "The URL of the service account impersonation request. Only with `external-oauth-wif`. It cannot be removed once set.",
					},
				},
			},
		},
	}
}
