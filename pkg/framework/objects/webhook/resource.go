package webhook

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &webhookResource{}
	_ resource.ResourceWithConfigure   = &webhookResource{}
	_ resource.ResourceWithImportState = &webhookResource{}
)

func WebhookResource() resource.Resource {
	return &webhookResource{}
}

type webhookResource struct {
	client *dbt_cloud.Client
}

func (r *webhookResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resourceSchema
}

func readWebhookToWebhookResourceModel(ctx context.Context, retrievedWebhook *dbt_cloud.WebhookRead, resourceModel *WebhookResourceModel) diag.Diagnostics {
	// these two are identical. WebhookID should not have been used in the
	// first place, but we're keeping it here for backwards compatibility reasons
	resourceModel.WebhookID = types.StringValue(retrievedWebhook.WebhookId)
	resourceModel.ID = types.StringValue(retrievedWebhook.WebhookId)

	resourceModel.Name = types.StringValue(retrievedWebhook.Name)
	resourceModel.Description = types.StringValue(retrievedWebhook.Description)
	resourceModel.ClientURL = types.StringValue(retrievedWebhook.ClientUrl)

	var diags diag.Diagnostics
	resourceModel.EventTypes, diags = helper.SliceStringToTypesListStringValue(retrievedWebhook.EventTypes)

	if diags.HasError() {
		return diags
	}
	resourceModel.JobIDs, diags = helper.SliceStringToTypesListInt64Value(retrievedWebhook.JobIds)
	if diags.HasError() {
		return diags
	}

	resourceModel.Active = types.BoolValue(retrievedWebhook.Active)

	resourceModel.HTTPStatusCode = types.StringValue(*retrievedWebhook.HttpStatusCode)
	resourceModel.AccountIdentifier = types.StringValue(*retrievedWebhook.AccountIdentifier)

	return nil
}

func (r *webhookResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state WebhookResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	webhookId := state.ID.ValueString()

	retrievedWebhook, err := r.client.GetWebhook(webhookId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The webhook was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting webhook", err.Error())
		return
	}
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = readWebhookToWebhookResourceModel(ctx, retrievedWebhook, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fmt.Printf("HTTP Status Code Type: %v\n", reflect.TypeOf(*retrievedWebhook.HttpStatusCode))
	fmt.Printf("Account Identifier Type: %v\n", reflect.TypeOf(*retrievedWebhook.AccountIdentifier))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	fmt.Printf("State after READ: %+v\n", state)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *webhookResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan WebhookResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	webhookId := ""
	name := plan.Name.ValueString()
	description := plan.Description.ValueString()
	clientUrl := plan.ClientURL.ValueString()
	eventTypes := plan.EventTypes
	jobIds := helper.TypesListInt64SliceToInt64Slice(plan.JobIDs)
	active := plan.Active.ValueBool()

	nonTypedEventTypes := helper.TypesListStringToStringSlice(eventTypes)

	createdWebhook, err := r.client.CreateWebhook(
		webhookId,
		name,
		description,
		clientUrl,
		nonTypedEventTypes,
		jobIds,
		active,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create webhook",
			"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(createdWebhook.WebhookId)
	plan.WebhookID = types.StringValue(createdWebhook.WebhookId)

	plan.JobIDs, diags = helper.SliceStringToTypesListInt64Value(createdWebhook.JobIds)
	if diags.HasError() {
		return
	}

	// Set computed fields
	plan.HmacSecret = types.StringValue(*createdWebhook.HmacSecret)
	plan.AccountIdentifier = types.StringValue(*createdWebhook.AccountIdentifier)
	plan.HTTPStatusCode = types.StringValue(*createdWebhook.HttpStatusCode)
	plan.Active = types.BoolValue(createdWebhook.Active)

	// Set the state with all fields
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	fmt.Printf("State after CREATE: %+v\n", plan)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *webhookResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state WebhookResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var nameChanged = (plan.Name != state.Name)
	var descriptionChanged = (plan.Description != state.Description)
	var clientUrlChanged = (plan.ClientURL != state.ClientURL)
	var eventTypesChanged = !reflect.DeepEqual(plan.EventTypes, state.EventTypes)
	var jobIdsChanged = !reflect.DeepEqual(plan.JobIDs, state.JobIDs)
	var activeChanged = !reflect.DeepEqual(plan.Active, state.Active)

	if nameChanged || descriptionChanged || clientUrlChanged || eventTypesChanged || jobIdsChanged || activeChanged {
		var webhookId = state.WebhookID.ValueString()
		retrievedWebhook, err := r.client.GetWebhook(webhookId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Issue getting webhook",
				"Error: "+err.Error(),
			)
			return
		}

		updateWebhook := dbt_cloud.WebhookWrite{
			WebhookId:   webhookId,
			Name:        helper.TernaryOperator(nameChanged, plan.Name.ValueString(), retrievedWebhook.Name),
			Description: helper.TernaryOperator(descriptionChanged, plan.Description.ValueString(), retrievedWebhook.Description),
			ClientUrl:   helper.TernaryOperator(clientUrlChanged, plan.ClientURL.ValueString(), retrievedWebhook.ClientUrl),
			EventTypes:  helper.TernaryOperator(eventTypesChanged, helper.TypesListStringToStringSlice(plan.EventTypes), retrievedWebhook.EventTypes),
			JobIds:      helper.TernaryOperator(jobIdsChanged, helper.TypesListInt64SliceToInt64Slice(plan.JobIDs), helper.SliceStringToSliceInt64(retrievedWebhook.JobIds)),
			Active:      helper.TernaryOperator(activeChanged, plan.Active.ValueBool(), retrievedWebhook.Active),
		}

		_, err = r.client.UpdateWebhook(webhookId, updateWebhook)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update webhook",
				"Error: "+err.Error(),
			)
			return
		}

		postUpdateRetrievedWebhook, err := r.client.GetWebhook(webhookId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Issue getting webhook post update",
				"Error: "+err.Error(),
			)
			return
		}
		diags = readWebhookToWebhookResourceModel(ctx, postUpdateRetrievedWebhook, &state)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	fmt.Printf("State after UPDATE: %+v\n", state)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *webhookResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state WebhookResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteWebhook(state.WebhookID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting webhook",
			"Error: "+err.Error(),
		)
		return
	}
}

func (r *webhookResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *webhookResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*dbt_cloud.Client)
}
