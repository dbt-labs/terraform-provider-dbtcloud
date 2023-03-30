package dbt_cloud

import (
	"context"
	"os"
    "strconv"

    old_dbt_cloud "github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ provider.Provider = &dbtCloudProvider{}
)

func New() provider.Provider {
	return &dbtCloudProvider{}
}

type dbtCloudProvider struct{}

func (p *dbtCloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dbt_cloud"
}

func (p *dbtCloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with dbt Cloud",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:    false,
				Sensitive:   true,
				Description: "API token for your dbt Cloud",
			},
			"account_id": schema.Int64Attribute{
				Optional:    false,
				Description: "Account identifier for your DBT Cloud implementation",
			},
			"host_url": schema.StringAttribute{
				Optional:    true,
				Description: "URL for your dbt Cloud deployment",
			},
		},
	}
}

type dbtCloudProviderModel struct {
	Token     types.String `tfsdk:"token"`
	AccountID types.Int64  `tfsdk:"account_id"`
	HostURL   types.String `tfsdk:"host_url"`
}

func (p *dbtCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring dbt Cloud client")

	var config dbtCloudProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.AccountID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("account_id"),
			"Unknown dbt Cloud account identifier",
			"dbt Cloud account identifier must be provided in order to establish a connection",
		)
	}

	if config.HostURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host_url"),
			"Unknown dbt Cloud host URL",
			"dbt Cloud host URL must be provided in order to establish a connection",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown dbt Cloud token",
			"Token must be provided in order to establish a connection",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	accountIDString := os.Getenv("DBT_CLOUD_ACCOUNT_ID")
	accountID, _ := strconv.Atoi(accountIDString)
	token := os.Getenv("DBT_CLOUD_TOKEN")
	hostURL := os.Getenv("DBT_CLOUD_HOST_URL")

	if !config.AccountID.IsNull() {
		accountID = int(config.AccountID.ValueInt64())
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if !config.HostURL.IsNull() {
		hostURL = config.HostURL.ValueString()
	}

	if accountID == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("account_id"),
			"Missing dbt Cloud account identifier",
			"dbt Cloud account identifier must be provided in order to establish a connection",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing dbt Cloud token",
			"dbt Cloud token must be provided in order to establish a connection, currently no default is set",
		)
	}

	if hostURL == "" {
		hostURL = "https://cloud.getdbt.com/api"
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := old_dbt_cloud.NewClient(&accountID, &token, &hostURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create dbt Cloud API Client",
			"An unexpected error occurred when creating the dbt Cloud API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"dbt Cloud API Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured dbt Cloud client", map[string]any{"success": true})
}

func (p *dbtCloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		UserDataSource,
	}
}

func (p *dbtCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}
