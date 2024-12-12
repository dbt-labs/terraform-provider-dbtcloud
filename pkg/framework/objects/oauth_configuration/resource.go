package oauth_configuration

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &oAuthConfigurationResource{}
	_ resource.ResourceWithConfigure      = &oAuthConfigurationResource{}
	_ resource.ResourceWithImportState    = &oAuthConfigurationResource{}
	_ resource.ResourceWithValidateConfig = &oAuthConfigurationResource{}
)

func OAuthConfigurationResource() resource.Resource {
	return &oAuthConfigurationResource{}
}

type oAuthConfigurationResource struct {
	client *dbt_cloud.Client
}

func (r *oAuthConfigurationResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_oauth_configuration"
}

func (r *oAuthConfigurationResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var data OAuthConfigurationResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Type.ValueString() == "okta" && !data.ApplicationIdUri.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("application_id_uri"),
			"application_id_uri is not supported for Okta",
			"application_id_uri is only supported for Entra ID (i.e. Azure AD) OAuth integrations",
		)
	}

	if data.Type.ValueString() == "entra" && data.ApplicationIdUri.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("application_id_uri"),
			"application_id_uri is required for Entra ID",
			"application_id_uri is required for Entra ID (i.e. Azure AD) OAuth integrations",
		)
	}
}

func (r *oAuthConfigurationResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state OAuthConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	oAuthConfigurationID := state.ID.ValueInt64()
	retrievedOAuthConfiguration, err := r.client.GetOAuthConfiguration(oAuthConfigurationID)

	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The OAuth configuration was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the OAuth configuration", err.Error())
		return
	}

	state.ID = types.Int64Value(int64(*retrievedOAuthConfiguration.ID))
	state.Type = types.StringValue(retrievedOAuthConfiguration.Type)
	state.Name = types.StringValue(retrievedOAuthConfiguration.Name)
	state.ClientId = types.StringValue(retrievedOAuthConfiguration.ClientId)
	state.AuthorizeUrl = types.StringValue(retrievedOAuthConfiguration.AuthorizeUrl)
	state.TokenUrl = types.StringValue(retrievedOAuthConfiguration.TokenUrl)
	state.RedirectUri = types.StringValue(retrievedOAuthConfiguration.RedirectUri)

	if retrievedOAuthConfiguration.OAuthConfigurationExtra != nil {
		state.ApplicationIdUri = types.StringValue(
			*retrievedOAuthConfiguration.OAuthConfigurationExtra.ApplicationIdUri,
		)
	} else {
		state.ApplicationIdUri = types.StringValue("")
	}

	// secrets are not set when reading.
	// Here the only secret is `client_secret`

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *oAuthConfigurationResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan OAuthConfigurationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oAuthType := plan.Type.ValueString()
	name := plan.Name.ValueString()
	clientID := plan.ClientId.ValueString()
	clientSecret := plan.ClientSecret.ValueString()
	authorizeURL := plan.AuthorizeUrl.ValueString()
	tokenURL := plan.TokenUrl.ValueString()
	redirectURI := plan.RedirectUri.ValueString()

	// will be set to "" if not configured
	applicationURI := plan.ApplicationIdUri.ValueString()

	createdOAuthConfiguration, err := r.client.CreateOAuthConfiguration(
		oAuthType,
		name,
		clientID,
		clientSecret,
		authorizeURL,
		tokenURL,
		redirectURI,
		applicationURI,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create OAuth configuration",
			"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.Int64Value(*createdOAuthConfiguration.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oAuthConfigurationResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state OAuthConfigurationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oAuthConfigurationID := state.ID.ValueInt64()

	err := r.client.DeleteOAuthConfiguration(oAuthConfigurationID)

	if err != nil {
		resp.Diagnostics.AddError(
			"Issue deleting OAuthConfiguration",
			"Error: "+err.Error(),
		)
		return
	}
}

func (r *oAuthConfigurationResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state OAuthConfigurationResourceModel

	// Read plan and state values into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	oAuthConfigurationID := state.ID.ValueInt64()

	retrievedOAuthConfiguration, err := r.client.GetOAuthConfiguration(oAuthConfigurationID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue getting OAuth configuration",
			"Error: "+err.Error(),
		)
		return
	}

	// we check the fields which don't trigger a new resource
	if plan.Name != state.Name {
		retrievedOAuthConfiguration.Name = plan.Name.ValueString()
	}
	if plan.ClientId != state.ClientId {
		retrievedOAuthConfiguration.ClientId = plan.ClientId.ValueString()
	}
	if plan.ClientSecret != state.ClientSecret {
		retrievedOAuthConfiguration.ClientSecret = plan.ClientSecret.ValueString()
	}
	if plan.AuthorizeUrl != state.AuthorizeUrl {
		retrievedOAuthConfiguration.AuthorizeUrl = plan.AuthorizeUrl.ValueString()
	}
	if plan.TokenUrl != state.TokenUrl {
		retrievedOAuthConfiguration.TokenUrl = plan.TokenUrl.ValueString()
	}
	if plan.RedirectUri != state.RedirectUri {
		retrievedOAuthConfiguration.RedirectUri = plan.RedirectUri.ValueString()
	}
	if plan.ApplicationIdUri != state.ApplicationIdUri {
		if plan.ApplicationIdUri.IsNull() {
			retrievedOAuthConfiguration.OAuthConfigurationExtra = nil
		} else {
			applicationIdUri := plan.ApplicationIdUri.ValueString()
			retrievedOAuthConfiguration.OAuthConfigurationExtra = &dbt_cloud.OAuthConfigurationExtra{
				ApplicationIdUri: &applicationIdUri,
			}
		}
	}

	_, err = r.client.UpdateOAuthConfiguration(
		oAuthConfigurationID,
		*retrievedOAuthConfiguration,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update OAuth configuration",
			"Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oAuthConfigurationResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {

	// I think we need this conversion because the ID is a string
	oAuthConfigurationIDStr := req.ID
	oAuthConfigurationID, err := strconv.Atoi(oAuthConfigurationIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing OAuth configuration ID for import", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("id"), oAuthConfigurationID,
	)...)

}

func (r *oAuthConfigurationResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
