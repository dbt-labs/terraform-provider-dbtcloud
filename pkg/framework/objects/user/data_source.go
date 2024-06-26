package user

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

func UserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *dbt_cloud.Client
}

type userDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
}

func (d *userDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve user details",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the user",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "Email for the user",
			},
		},
	}
}

func (d *userDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state userDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	user, err := d.client.GetUser(string(state.Email.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Did not find user with email: %s", state.Email.ValueString()),
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

func (d *userDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}
