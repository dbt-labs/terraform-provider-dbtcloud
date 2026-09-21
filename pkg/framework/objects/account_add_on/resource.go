package account_add_on

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &accountAddOnResource{}
	_ resource.ResourceWithConfigure   = &accountAddOnResource{}
	_ resource.ResourceWithImportState = &accountAddOnResource{}
)

type accountAddOnResource struct {
	client *dbt_cloud.Client
}

func AccountAddOnResource() resource.Resource {
	return &accountAddOnResource{}
}

func (r *accountAddOnResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_account_add_on"
}

// productUnavailableError is raised when the account cannot use the product at all.
// The endpoint leaves such a product out of its response, so there is nothing to read.
func productUnavailableError(product string) (string, string) {
	return "Add-on product not available",
		fmt.Sprintf(
			"The account cannot use the %q add-on, so its state cannot be read. "+
				"`wizard` needs the AI features of the account to be turned on, which "+
				"`dbtcloud_account_features` manages. Run `terraform state rm` for this "+
				"resource if the product is not meant to be managed here.",
			product,
		)
}

// readAddOn collects the state of one product. It returns a nil model when the
// account has never used the product, which means that the resource is gone.
func readAddOn(client *dbt_cloud.Client, product string) (*AccountAddOnResourceModel, bool, error) {
	addOn, err := client.GetAccountAddOn(product)
	if err != nil {
		return nil, false, err
	}
	if addOn == nil {
		return nil, true, nil
	}
	// No state means that the account has never activated or trialled the product.
	if addOn.State == nil {
		return nil, false, nil
	}

	spendLimit, err := client.GetAccountAddOnSpendLimit(product)
	if err != nil {
		return nil, false, err
	}

	model := AccountAddOnResourceModel{
		ID:             types.StringValue(product),
		Product:        types.StringValue(product),
		State:          types.StringPointerValue(addOn.State),
		TrialStartedAt: types.StringPointerValue(addOn.TrialStartedAt),
		TrialEndsAt:    types.StringPointerValue(addOn.TrialEndsAt),
		TrialConsumed:  types.BoolValue(addOn.TrialConsumed),
		CanActivate:    types.BoolValue(addOn.CanActivate),
		CanTrial:       types.BoolValue(addOn.CanTrial),
	}
	if spendLimit != nil {
		model.SpendLimitNanodollars = types.Int64Value(*spendLimit)
	} else {
		model.SpendLimitNanodollars = types.Int64Null()
	}

	return &model, false, nil
}

func spendLimitPointer(value types.Int64) *int64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	limit := value.ValueInt64()
	return &limit
}

func (r *accountAddOnResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan AccountAddOnResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	product := plan.Product.ValueString()

	var err error
	if plan.Activation.ValueString() == activationTrial {
		_, err = r.client.StartAccountAddOnTrial(product)
	} else {
		_, err = r.client.ActivateAccountAddOn(product)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error turning on the %s add-on", product),
			err.Error(),
		)
		return
	}

	// The add-on is on from here. Any later failure still records the resource, so
	// that a second apply does not leave an add-on nobody tracks.
	if limit := spendLimitPointer(plan.SpendLimitNanodollars); limit != nil {
		if _, err := r.client.SetAccountAddOnSpendLimit(product, limit); err != nil {
			r.saveAfterPartialCreate(ctx, product, plan, resp)
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error setting the spend limit of the %s add-on", product),
				"The add-on is on, but its spend limit was not set: "+err.Error(),
			)
			return
		}
	}

	model, unavailable, err := readAddOn(r.client, product)
	if err != nil {
		r.saveAfterPartialCreate(ctx, product, plan, resp)
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading the %s add-on", product),
			err.Error(),
		)
		return
	}
	if unavailable {
		summary, detail := productUnavailableError(product)
		resp.Diagnostics.AddError(summary, detail)
		return
	}
	if model == nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading the %s add-on", product),
			"The add-on was turned on, but the account still reports no state for it.",
		)
		return
	}

	model.Activation = plan.Activation
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

// saveAfterPartialCreate records the resource when the add-on was turned on but a
// later call failed, so that the add-on stays tracked.
func (r *accountAddOnResource) saveAfterPartialCreate(
	ctx context.Context,
	product string,
	plan AccountAddOnResourceModel,
	resp *resource.CreateResponse,
) {
	model, unavailable, err := readAddOn(r.client, product)
	if err != nil || unavailable || model == nil {
		return
	}
	model.Activation = plan.Activation
	resp.State.Set(ctx, model)
}

func (r *accountAddOnResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state AccountAddOnResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	product := state.Product.ValueString()

	model, unavailable, err := readAddOn(r.client, product)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading the %s add-on", product),
			err.Error(),
		)
		return
	}
	if unavailable {
		summary, detail := productUnavailableError(product)
		resp.Diagnostics.AddError(summary, detail)
		return
	}
	// The account no longer holds the add-on, so the resource is gone.
	if model == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	model.Activation = state.Activation
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *accountAddOnResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state AccountAddOnResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	product := plan.Product.ValueString()

	// product and activation both replace the resource, so the spend limit is the
	// only value an update can change.
	if !plan.SpendLimitNanodollars.Equal(state.SpendLimitNanodollars) {
		_, err := r.client.SetAccountAddOnSpendLimit(
			product,
			spendLimitPointer(plan.SpendLimitNanodollars),
		)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error setting the spend limit of the %s add-on", product),
				err.Error(),
			)
			return
		}
	}

	model, unavailable, err := readAddOn(r.client, product)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading the %s add-on", product),
			err.Error(),
		)
		return
	}
	if unavailable {
		summary, detail := productUnavailableError(product)
		resp.Diagnostics.AddError(summary, detail)
		return
	}
	if model == nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading the %s add-on", product),
			"The account reports no state for the add-on after the update.",
		)
		return
	}

	model.Activation = plan.Activation
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

func (r *accountAddOnResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state AccountAddOnResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	product := state.Product.ValueString()

	if _, err := r.client.CancelAccountAddOn(product); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error cancelling the %s add-on", product),
			"An add-on that a contract pays for cannot be cancelled with a token, and "+
				"needs dbt Labs support. "+err.Error(),
		)
		return
	}
}

func (r *accountAddOnResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	product := req.ID

	// activation only picks the call that turns the product on, and the API does not
	// report which one was used. It still has to be filled in, because an empty value
	// reads as a change against the default and asks to replace the resource right
	// after the import. A product in its trial was turned on as a trial.
	activation := activationPaid
	addOn, err := r.client.GetAccountAddOn(product)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error importing the %s add-on", product),
			err.Error(),
		)
		return
	}
	if addOn == nil {
		summary, detail := productUnavailableError(product)
		resp.Diagnostics.AddError(summary, detail)
		return
	}
	if addOn.State != nil && *addOn.State == dbt_cloud.AddOnStateTrial {
		activation = activationTrial
	}

	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("product"), product)...,
	)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("id"), product)...,
	)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("activation"), activation)...,
	)
}

func (r *accountAddOnResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
