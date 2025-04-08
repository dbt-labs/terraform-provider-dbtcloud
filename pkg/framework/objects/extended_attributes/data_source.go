package extended_attributes

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &extendedAttributesDataSource{}
	_ datasource.DataSourceWithConfigure = &extendedAttributesDataSource{}
)

func ExtendedAttributesDataSource() datasource.DataSource {
	return &extendedAttributesDataSource{}
}

type extendedAttributesDataSource struct {
	client *dbt_cloud.Client
}

func (p *extendedAttributesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dbt_cloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *dbt_cloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	p.client = client
}

func (p *extendedAttributesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extended_attributes"
}

func (p *extendedAttributesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ExtendedAttributesDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectId := int(state.ProjectID.ValueInt64())
	extendedAttributesId := int(state.ExtendedAttributesID.ValueInt64())

	extendedAttributes, err := p.client.GetExtendedAttributes(projectId, extendedAttributesId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Extended attributes",
			"Could not read Extended attributes ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d%s%d", *extendedAttributes.ID, dbt_cloud.ID_DELIMITER, *extendedAttributes.ID))
	state.ProjectID = types.Int64Value(int64(extendedAttributes.ProjectID))
	state.ExtendedAttributesID = types.Int64Value(int64(*extendedAttributes.ID))
	state.ExtendedAttributes = types.StringValue(string(extendedAttributes.ExtendedAttributes))
	state.State = types.Int64Value(int64(extendedAttributes.State))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *extendedAttributesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}
