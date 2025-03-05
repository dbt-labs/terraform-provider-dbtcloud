package model_notifications

import (
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ModelNotificationsResourceModel struct {
	ID            types.String `tfsdk:"id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	OnSuccess     types.Bool   `tfsdk:"on_success"`
	OnFailure     types.Bool   `tfsdk:"on_failure"`
	OnWarning     types.Bool   `tfsdk:"on_warning"`
	OnSkipped     types.Bool   `tfsdk:"on_skipped"`
}

type ModelNotificationsDataSourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	OnSuccess     types.Bool   `tfsdk:"on_success"`
	OnFailure     types.Bool   `tfsdk:"on_failure"`
	OnWarning     types.Bool   `tfsdk:"on_warning"`
	OnSkipped     types.Bool   `tfsdk:"on_skipped"`
}

func ConvertModelNotificationsModelToData(model ModelNotificationsResourceModel) dbt_cloud.ModelNotifications {
	environmentID, _ := strconv.Atoi(model.EnvironmentID.ValueString())

	modelNotifications := dbt_cloud.ModelNotifications{
		EnvironmentID: environmentID,
		Enabled:       model.Enabled.ValueBool(),
		OnSuccess:     model.OnSuccess.ValueBool(),
		OnFailure:     model.OnFailure.ValueBool(),
		OnWarning:     model.OnWarning.ValueBool(),
		OnSkipped:     model.OnSkipped.ValueBool(),
	}

	if !model.ID.IsNull() {
		idStr := model.ID.ValueString()
		id, err := strconv.Atoi(idStr)
		if err == nil {
			modelNotifications.ID = &id
		}
	}

	return modelNotifications
}
