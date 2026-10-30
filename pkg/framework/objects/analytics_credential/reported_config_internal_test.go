package analytics_credential

import (
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Rule: the non-secret fields in the state are the ones the API reports after an update.
// given: a planned Snowflake block and an update response that changes the warehouse and leaves the user out
// when: the reported configuration is applied
// then: the warehouse is the reported one and the user keeps the planned value.
func TestApplyReportedConfig_TakesReportedFieldsAndKeepsOmittedOnes(t *testing.T) {
	model := &AnalyticsCredentialResourceModel{Snowflake: &SnowflakeModel{
		User:      types.StringValue("planned-user"),
		Warehouse: types.StringValue("PLANNED_WH"),
	}}
	credential := &dbt_cloud.AnalyticsCredential{Config: dbt_cloud.AnalyticsCredentialConfig{Warehouse: "REPORTED_WH"}}

	applyReportedConfig(model, credential)

	if got := model.Snowflake.Warehouse.ValueString(); got != "REPORTED_WH" {
		t.Fatalf("expected the reported warehouse, got %q", got)
	}
	if got := model.Snowflake.User.ValueString(); got != "planned-user" {
		t.Fatalf("expected the planned user to be kept, got %q", got)
	}
}

// Rule: a field left out of the configuration stays unset, because the server may fill in a default.
// given: a planned BigQuery block without auth_type and an update response that reports service-account-json
// when: the reported configuration is applied
// then: auth_type stays null.
func TestApplyReportedConfig_KeepsUnconfiguredFieldsNull(t *testing.T) {
	model := &AnalyticsCredentialResourceModel{BigQuery: &BigQueryModel{
		AuthType:         types.StringNull(),
		ExecutionProject: types.StringValue("planned-project"),
	}}
	credential := &dbt_cloud.AnalyticsCredential{Config: dbt_cloud.AnalyticsCredentialConfig{
		AuthType:         "service-account-json",
		ExecutionProject: "reported-project",
	}}

	applyReportedConfig(model, credential)

	if !model.BigQuery.AuthType.IsNull() {
		t.Fatalf("expected auth_type to stay null, got %v", model.BigQuery.AuthType)
	}
	if got := model.BigQuery.ExecutionProject.ValueString(); got != "reported-project" {
		t.Fatalf("expected the reported execution project, got %q", got)
	}
}

// Rule: the adapter version in the state is the one the API reports, with a fallback when it is left out.
// given: a reported adapter version, an empty one with a known fallback, and an empty one with no fallback
// when: the adapter version is resolved
// then: the reported value, the fallback, and null are returned.
func TestReportedAdapterVersion(t *testing.T) {
	cases := map[string]struct {
		fallback types.String
		reported string
		want     types.String
	}{
		"reported":    {types.StringValue("snowflake_v0"), "databricks_v0", types.StringValue("databricks_v0")},
		"fallback":    {types.StringValue("snowflake_v0"), "", types.StringValue("snowflake_v0")},
		"nothing":     {types.StringValue(""), "", types.StringNull()},
		"unknown one": {types.StringUnknown(), "", types.StringNull()},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := reportedAdapterVersion(tc.fallback, tc.reported); !got.Equal(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
