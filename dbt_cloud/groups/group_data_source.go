package dbt_cloud

import (
	"context"

	dbt_cloud_old "github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

func GroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

type groupDataSource struct {
	client *dbt_cloud_old.Client
}

type groupDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	IsActive types.Boolean `tfsdk:"is_active"`
	AssignByDefault types.Boolean `tfsdk:"assign_by_default"`
	SSOMappingGroups []types.String `tfsdk:"sso_mapping_groups"`
}

func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieve group details",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:    true,
				Description: "ID of the user",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name for the group",
			},
		},
	}
}

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	user, err := d.client.GetUser(string(state.Email.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read dbt Cloud User",
			err.Error(),
		)
		return
	}

	state.ID = types.Int64Value(int64(user.ID))
	state.Email = types.StringValue(user.Email)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud_old.Client)
}
