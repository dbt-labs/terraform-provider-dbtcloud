package environment

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &environmentDataSource{}
	_ datasource.DataSourceWithConfigure = &environmentDataSource{}
)

func EnvironmentDataSource() datasource.DataSource {
	return &environmentDataSource{}
}

type environmentDataSource struct {
	client *dbt_cloud.Client
}

func (d *environmentDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *environmentDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var config EnvironmentDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	environment, err := d.client.GetEnvironment(
		int(config.ProjectID.ValueInt64()),
		int(config.EnvironmentID.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Did not find environment with this ID: %s", config.EnvironmentID),
			err.Error(),
		)
		return
	}

	state := config

	state.CredentialsID = types.Int64PointerValue(
		helper.IntPointerToInt64Pointer(environment.Credential_Id),
	)
	state.Name = types.StringValue(environment.Name)
	state.DbtVersion = types.StringValue(environment.Dbt_Version)
	state.Type = types.StringValue(environment.Type)
	state.UseCustomBranch = types.BoolValue(environment.Use_Custom_Branch)
	state.CustomBranch = types.StringPointerValue(environment.Custom_Branch)
	state.DeploymentType = types.StringPointerValue(environment.DeploymentType)
	state.ExtendedAttributesID = types.Int64PointerValue(
		helper.IntPointerToInt64Pointer(environment.ExtendedAttributesID),
	)
	state.EnableModelQueryHistory = types.BoolValue(environment.EnableModelQueryHistory)
	state.PrimaryProfileID = types.Int64PointerValue(
		helper.IntPointerToInt64Pointer(environment.PrimaryProfileID),
	)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *environmentDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}
