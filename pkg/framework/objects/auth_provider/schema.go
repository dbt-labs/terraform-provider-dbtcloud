package auth_provider

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var validAuthProviderTypes = []string{
	"saml",
	"okta",
	"gsuite",
	"azure_single_tenant",
	"azure_multi_tenant",
	"azure_active_directory",
}

func (r *authProviderResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resource_schema.Schema{
		Description: helper.DocString(
			`Manages an SSO auth provider for a dbt Cloud account. Supports SAML/Okta, Azure Active Directory (single-tenant, multi-tenant), and Google Workspace.

			Only one auth provider may exist per account. Requires the SSO feature enabled on the account (enterprise plans only).

			See the [documentation](https://docs.getdbt.com/docs/cloud/manage-access/sso-overview) for more information.`,
		),
		Attributes: map[string]resource_schema.Attribute{

			// ── Computed / read-only ──────────────────────────────────────
			"id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the auth provider.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"state": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "The state of the auth provider (1 = active).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"login_url": resource_schema.StringAttribute{
				Computed:    true,
				Description: "The SSO login URL for the account, auto-generated from the slug.",
			},
			"cert_expiry_date": resource_schema.StringAttribute{
				Computed:    true,
				Description: "Expiry date of the SAML X.509 certificate (SAML/Okta only).",
			},
			"created_at": resource_schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": resource_schema.StringAttribute{
				Computed: true,
			},

			// ── Common ────────────────────────────────────────────────────
			"type": resource_schema.StringAttribute{
				Required: true,
				Description: "The SSO provider type. One of: `saml`, `okta`, `gsuite`, `azure_single_tenant`, `azure_multi_tenant`, `azure_active_directory`. Changing this value forces a new resource.",
				Validators: []validator.String{
					stringvalidator.OneOf(validAuthProviderTypes...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"slug": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL-safe identifier used in the SSO login URL. Auto-generated if omitted. Immutable on accounts where auto-slug enforcement is enabled.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_password_backdoor": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "When true (default), users can still log in with email and password as a fallback. Set to false to enforce SSO-only access.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// ── SAML / Okta ───────────────────────────────────────────────
			"entity_id": resource_schema.StringAttribute{
				Optional:    true,
				Description: "SAML entity ID (Issuer) from your identity provider. Required for `saml` and `okta`.",
			},
			"sso_url": resource_schema.StringAttribute{
				Optional:    true,
				Description: "SAML Single Sign-On URL from your identity provider. Required for `saml` and `okta`.",
			},
			"cert": resource_schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "SAML X.509 certificate (PEM format). Sensitive — stored in state. Consider using `cert_wo` instead. Conflicts with `cert_wo`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("cert_wo")),
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("cert_wo")),
				},
			},
			"cert_wo": resource_schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Description: "Write-only alternative to `cert`. Not stored in state. Use `cert_wo_version` to trigger updates. Conflicts with `cert`.",
			},
			"cert_wo_version": resource_schema.Int64Attribute{
				Optional:    true,
				Description: "Increment to rotate `cert_wo` without changing the value.",
			},
			"sign_request": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to sign SAML authentication requests. Defaults to false.",
			},
			"attribute_map": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("{}"),
				Description: "JSON map of SAML attribute names to dbt Cloud user fields.",
			},

			// ── Azure AD / Google Workspace ───────────────────────────────
			"client_id": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "OAuth client ID. Required for Azure AD and Google Workspace providers. Not returned by the API after save (encrypted at rest).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_secret": resource_schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "OAuth client secret. Required for Azure AD and Google Workspace providers. Sensitive — stored in state. Consider using `client_secret_wo` instead. Conflicts with `client_secret_wo`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("client_secret_wo")),
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("client_secret_wo")),
				},
			},
			"client_secret_wo": resource_schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Description: "Write-only alternative to `client_secret`. Not stored in state. Use `client_secret_wo_version` to trigger updates. Conflicts with `client_secret`.",
			},
			"client_secret_wo_version": resource_schema.Int64Attribute{
				Optional:    true,
				Description: "Increment to rotate `client_secret_wo` without changing the value.",
			},

			// ── Azure AD only ─────────────────────────────────────────────
			"tenant_id": resource_schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Azure AD tenant ID. Required for `azure_single_tenant`.",
			},
			"domain": resource_schema.StringAttribute{
				Optional:    true,
				Description: "Primary domain for the Azure AD or Google Workspace tenant.",
			},
			"include_indirect_groups": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to include transitive (indirect) group memberships from Azure AD. Defaults to true.",
			},
			"max_groups_to_retrieve": resource_schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(250),
				Description: "Maximum number of Azure AD groups to fetch per user. Defaults to 250.",
			},

			// ── Google Workspace only ─────────────────────────────────────
			"gsuite_admin_id": resource_schema.StringAttribute{
				Optional:    true,
				Description: "Google Workspace admin email used to fetch group memberships.",
			},
			"authorization_url": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "OAuth authorization URL for Google Workspace. May be auto-populated server-side.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"admin_refresh_token": resource_schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Google Workspace admin OAuth refresh token used to fetch group memberships.",
			},
			"resource_metadata": resource_schema.DynamicAttribute{
				Optional:    true,
				Description: "Metadata for tracking resource identity during account migrations. Stored in Terraform state only and not sent to the API.",
			},
		},
	}
}
