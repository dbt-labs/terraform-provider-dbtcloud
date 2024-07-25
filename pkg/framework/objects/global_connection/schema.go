package global_connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *globalConnectionResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"adapter_version": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"is_ssh_tunnel_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"private_link_endpoint_id": schema.Int64Attribute{
				Optional: true,
			},
			"oauth_configuration_id": schema.Int64Attribute{
				Optional: true,
			},
			"bigquery": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"gcp_project_id": schema.StringAttribute{
						Required: true,
					},
					"timeout_seconds": schema.Int64Attribute{
						Required: true,
					},
					"private_key_id": schema.StringAttribute{
						Required: true,
					},
					"private_key": schema.StringAttribute{
						Required: true,
					},
					"client_email": schema.StringAttribute{
						Required: true,
					},
					"client_id": schema.StringAttribute{
						Required: true,
					},
					"auth_uri": schema.StringAttribute{
						Required: true,
					},
					"token_uri": schema.StringAttribute{
						Required: true,
					},
					"auth_provider_x509_cert_url": schema.StringAttribute{
						Required: true,
					},
					"client_x509_cert_url": schema.StringAttribute{
						Required: true,
					},
					"priority": schema.StringAttribute{
						Optional: true,
					},
					"retries": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Default:  int64default.StaticInt64(1),
					},
					"location": schema.StringAttribute{
						Optional: true,
					},
					"maximum_bytes_billed": schema.Int64Attribute{
						Optional: true,
					},
					"execution_project": schema.StringAttribute{
						Optional: true,
					},
					"impersonate_service_account": schema.StringAttribute{
						Optional: true,
					},
					"job_retry_deadline_seconds": schema.Int64Attribute{
						Optional: true,
					},
					"job_creation_timeout_seconds": schema.Int64Attribute{
						Optional: true,
					},
					"application_id": schema.StringAttribute{
						Required:    true,
						Description: "OAuth Client ID",
					},
					"application_secret": schema.StringAttribute{
						Required:    true,
						Description: "OAuth Client Secret",
					},
					"gcs_bucket": schema.StringAttribute{
						Optional: true,
					},
					"dataproc_region": schema.StringAttribute{
						Optional: true,
					},
					"dataproc_cluster_name": schema.StringAttribute{
						Optional: true,
					},
					"scopes": schema.SetAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default: setdefault.StaticValue(
							types.SetValueMust(
								types.StringType,
								[]attr.Value{
									types.StringValue("https://www.googleapis.com/auth/bigquery"),
									types.StringValue(
										"https://www.googleapis.com/auth/cloud-platform",
									),
									types.StringValue("https://www.googleapis.com/auth/drive"),
								},
							),
						),
					},
				},
			},
			// this feels bad, but there is no error/warning when people add extra fields https://github.com/hashicorp/terraform/issues/33570
			"snowflake": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"account": schema.StringAttribute{
						Required: true,
					},
					"database": schema.StringAttribute{
						Required: true,
					},
					"warehouse": schema.StringAttribute{
						Required: true,
					},
					"allow_sso": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
					// TODO: required if allow_sso is true
					"oauth_client_id": schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
					},
					"oauth_client_secret": schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
					},
					"role": schema.StringAttribute{
						Optional: true,
					},
					"client_session_keep_alive": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(false),
					},
				},
			},
		},
	}
}
