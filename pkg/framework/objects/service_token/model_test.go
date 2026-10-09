package service_token_test

import (
	"context"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/service_token"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func categories(values ...string) types.Set {
	elements := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	return types.SetValueMust(types.StringType, elements)
}

// An empty list means no environment write access, so "all" has to reach the
// request rather than being dropped.
func TestPermissionRequestKeepsAll(t *testing.T) {
	permissions := []service_token.ServiceTokenPermission{
		{
			PermissionSet:                 types.StringValue("developer"),
			AllProjects:                   types.BoolValue(true),
			WritableEnvironmentCategories: categories("all"),
		},
	}

	data, diags := service_token.ConvertServiceTokenPermissionModelToData(
		context.Background(), permissions, 1, 2,
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}
	if got := data[0].WritableEnvs; len(got) != 1 || got[0] != dbt_cloud.EnvironmentCategory_All {
		t.Errorf("WritableEnvs = %v, want [all]", got)
	}
}

func TestPermissionRequestKeepsNamedCategories(t *testing.T) {
	permissions := []service_token.ServiceTokenPermission{
		{
			PermissionSet:                 types.StringValue("developer"),
			AllProjects:                   types.BoolValue(true),
			WritableEnvironmentCategories: categories("development", "staging"),
		},
	}

	data, diags := service_token.ConvertServiceTokenPermissionModelToData(
		context.Background(), permissions, 1, 2,
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}
	if len(data[0].WritableEnvs) != 2 {
		t.Errorf("WritableEnvs = %v, want two categories", data[0].WritableEnvs)
	}
}

// An empty list from the API has to stay empty, otherwise a token with no
// environment write access reads back as if it had all of them.
func TestPermissionModelKeepsEmptyList(t *testing.T) {
	data := []dbt_cloud.ServiceTokenPermission{
		{Set: "developer", AllProjects: true, WritableEnvs: []dbt_cloud.EnvironmentCategory{}},
	}

	model, diags := service_token.ConvertServiceTokenPermissionDataToModel(
		context.Background(), data,
	)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}
	if got := len(model[0].WritableEnvironmentCategories.Elements()); got != 0 {
		t.Errorf("WritableEnvironmentCategories has %d elements, want 0", got)
	}
}
