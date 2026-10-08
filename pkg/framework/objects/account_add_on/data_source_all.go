package account_add_on

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &accountAddOnsDataSource{}
	_ datasource.DataSourceWithConfigure = &accountAddOnsDataSource{}
)

type accountAddOnsDataSource struct {
	client *dbt_cloud.Client
}

func AccountAddOnsDataSource() datasource.DataSource {
	return &accountAddOnsDataSource{}
}

func (d *accountAddOnsDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_account_add_ons"
}

func (d *accountAddOnsDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state AccountAddOnsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addOns, err := d.client.GetAccountAddOns()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading the account add-ons",
			err.Error(),
		)
		return
	}

	allAddOns := []AccountAddOnDataSourceModel{}
	for _, addOn := range addOns {
		model, err := addOnDataSourceModel(d.client, addOn)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading the spend limit of the "+addOn.Product+" add-on",
				err.Error(),
			)
			return
		}
		allAddOns = append(allAddOns, model)
	}

	state.AddOns = allAddOns

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *accountAddOnsDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}
