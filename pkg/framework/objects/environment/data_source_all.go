package environment

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &environmentsDataSources{}
	_ datasource.DataSourceWithConfigure = &environmentsDataSources{}
)

func EnvironmentsDataSource() datasource.DataSource {
	return &environmentsDataSources{}
}

type environmentsDataSources struct {
	client *dbt_cloud.Client
}

func (d *environmentsDataSources) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

func (d *environmentsDataSources) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var config EnvironmentsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	var projectID int
	if config.ProjectID.IsNull() {
		projectID = 0
	} else {
		projectID = int(config.ProjectID.ValueInt64())
	}

	environments, err := d.client.GetAllEnvironments(projectID)

	if err != nil {
		resp.Diagnostics.AddError(
			"Issue when retrieving environments",
			err.Error(),
		)
		return
	}

	state := config

	allEnvs := []EnvironmentDataSourceModel{}
	for _, environment := range environments {
		currentEnv := EnvironmentDataSourceModel{}

		currentEnv.EnvironmentID = types.Int64PointerValue(
			helper.IntPointerToInt64Pointer(environment.Environment_Id),
		)
		currentEnv.ProjectID = types.Int64Value(int64(environment.Project_Id))

		currentEnv.CredentialsID = types.Int64PointerValue(
			helper.IntPointerToInt64Pointer(environment.Credential_Id),
		)
		currentEnv.Name = types.StringValue(environment.Name)
		currentEnv.DbtVersion = types.StringValue(environment.Dbt_Version)
		currentEnv.Type = types.StringValue(environment.Type)
		currentEnv.UseCustomBranch = types.BoolValue(environment.Use_Custom_Branch)
		currentEnv.CustomBranch = types.StringPointerValue(environment.Custom_Branch)
		currentEnv.DeploymentType = types.StringPointerValue(environment.DeploymentType)
		currentEnv.ExtendedAttributesID = types.Int64PointerValue(
			helper.IntPointerToInt64Pointer(environment.ExtendedAttributesID),
		)
		allEnvs = append(allEnvs, currentEnv)
	}
	state.Environments = allEnvs

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *environmentsDataSources) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}
