package service_token

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &serviceTokenDataSource{}
	_ datasource.DataSourceWithConfigure = &serviceTokenDataSource{}
)

func ServiceTokenDataSource() datasource.DataSource {
	return &serviceTokenDataSource{}
}

type serviceTokenDataSource struct {
	client *dbt_cloud.Client
}

// Metadata implements datasource.DataSource.
func (st *serviceTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_token"
}

// Configure implements datasource.DataSourceWithConfigure.
func (st *serviceTokenDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dbt_cloud.Client)

	if !ok {
		resp.Diagnostics.AddError("Only Failing on CI??", fmt.Sprintf("Failed to get client from provider data:\n\ttype: %v\n\tvalue: %v", reflect.TypeOf(req.ProviderData), req.ProviderData))
		resp.Diagnostics.AddError("Missing client", "A client is required to configure the service token resource")
		return
	}

	st.client = client
}

// Schema implements datasource.DataSource.
func (st *serviceTokenDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_token_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the service token",
				// TODO(cwalden): Deprecate this in favor of `id`
				// DeprecationMessage: "Use `id` instead",
			},
			// TODO(cwalden): use Int64Attribute after deprecating `service_token_id`
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service token",
			},
			"uid": schema.StringAttribute{
				Description: "Service token UID (part of the token)",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Service token name",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"service_token_permissions": schema.SetNestedBlock{
				Description: "Permissions set for the service token",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"permission_set": schema.StringAttribute{
							Description: "Set of permissions to apply",
							Computed:    true,
						},
						"all_projects": schema.BoolAttribute{
							Description: "Whether or not to apply this permission to all projects for this service token",
							Computed:    true,
						},
						// TODO(cwalden): Would this be better as a Set of Int64?
						// TODO(cwalden): Add a validator to ensure that the project ID is set if `all_projects` is false
						"project_id": schema.Int64Attribute{
							Description: "Project ID to apply this permission to for this service token",
							Computed:    true,
						},
						// TODO(cwalden): Add validator to ensure that this is configurable for the given `permission_set`
						"writable_environment_categories": schema.SetAttribute{
							Description: helper.DocString(
								`What types of environments to apply Write permissions to.
								Even if Write access is restricted to some environment types, the permission set will have Read access to all environments.
								The values allowed are ~~~all~~~, ~~~development~~~, ~~~staging~~~, ~~~production~~~ and ~~~other~~~.
								Not setting a value is the same as selecting ~~~all~~~.
								Not all permission sets support environment level write settings, only ~~~analyst~~~, ~~~database_admin~~~, ~~~developer~~~, ~~~git_admin~~~ and ~~~team_admin~~~.`,
							),
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Read implements datasource.DataSource.
func (st *serviceTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var data ServiceTokenDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svcTokID := int(data.ServiceTokenID.ValueInt64())

	svcTok, err := st.client.GetServiceToken(svcTokID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting the service token", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.Itoa(*svcTok.ID))
	data.Name = types.StringValue(svcTok.Name)
	data.UID = types.StringValue(svcTok.UID)

	svcTokPermissions, diags := ConvertServiceTokenPermissionDataToModel(ctx, svcTok.Permissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ServiceTokenPermissions = svcTokPermissions

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
