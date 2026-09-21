package account_add_on

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &accountAddOnDataSource{}
	_ datasource.DataSourceWithConfigure = &accountAddOnDataSource{}
)

type accountAddOnDataSource struct {
	client *dbt_cloud.Client
}

func AccountAddOnDataSource() datasource.DataSource {
	return &accountAddOnDataSource{}
}

func (d *accountAddOnDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_account_add_on"
}

// addOnDataSourceModel maps one add-on, and reads its spend limit when the account
// holds the product. A product with no state has no spend limit to read.
func addOnDataSourceModel(
	client *dbt_cloud.Client,
	addOn dbt_cloud.AccountAddOn,
) (AccountAddOnDataSourceModel, error) {
	model := AccountAddOnDataSourceModel{
		ID:                    types.StringValue(addOn.Product),
		Product:               types.StringValue(addOn.Product),
		State:                 types.StringPointerValue(addOn.State),
		TrialStartedAt:        types.StringPointerValue(addOn.TrialStartedAt),
		TrialEndsAt:           types.StringPointerValue(addOn.TrialEndsAt),
		TrialConsumed:         types.BoolValue(addOn.TrialConsumed),
		CanActivate:           types.BoolValue(addOn.CanActivate),
		CanTrial:              types.BoolValue(addOn.CanTrial),
		SpendLimitNanodollars: types.Int64Null(),
	}

	if addOn.State == nil {
		return model, nil
	}

	spendLimit, err := client.GetAccountAddOnSpendLimit(addOn.Product)
	if err != nil {
		return model, err
	}
	if spendLimit != nil {
		model.SpendLimitNanodollars = types.Int64Value(*spendLimit)
	}

	return model, nil
}

func (d *accountAddOnDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var config AccountAddOnDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	product := config.Product.ValueString()

	addOn, err := d.client.GetAccountAddOn(product)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading the %s add-on", product),
			err.Error(),
		)
		return
	}
	if addOn == nil {
		summary, detail := productUnavailableError(product)
		resp.Diagnostics.AddError(summary, detail)
		return
	}

	model, err := addOnDataSourceModel(d.client, *addOn)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading the spend limit of the %s add-on", product),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (d *accountAddOnDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*dbt_cloud.Client)
}
