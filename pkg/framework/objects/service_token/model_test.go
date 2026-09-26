package service_token

import (
	"context"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServiceTokenWritableEnvsRoundTrip guards against
// https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/731 where
// setting writable_environment_categories = ["all"] resulted in a token
// with no environment write access at all, and the drift was then hidden
// because an empty response from the API was rewritten to ["all"] in state.
func TestServiceTokenWritableEnvsRoundTrip(t *testing.T) {
	ctx := context.Background()

	t.Run("configuring all sends all in the request", func(t *testing.T) {
		allSet := types.SetValueMust(types.StringType, []attr.Value{
			types.StringValue(dbt_cloud.EnvironmentCategory_All),
		})

		permissions := []ServiceTokenPermission{
			{
				PermissionSet:                 types.StringValue("developer"),
				AllProjects:                   types.BoolValue(true),
				WritableEnvironmentCategories: allSet,
			},
		}

		requests, diags := ConvertServiceTokenPermissionModelToData(ctx, permissions, 1, 2)

		require.False(t, diags.HasError())
		require.Len(t, requests, 1)
		assert.ElementsMatch(
			t,
			[]dbt_cloud.EnvironmentCategory{dbt_cloud.EnvironmentCategory_All},
			requests[0].WritableEnvs,
			"the request must carry the writable environment categories the user configured, not drop them",
		)
	})

	t.Run("an empty response from the API is not rewritten to all", func(t *testing.T) {
		apiPermissions := []dbt_cloud.ServiceTokenPermission{
			{
				Set:          "developer",
				AllProjects:  true,
				WritableEnvs: []dbt_cloud.EnvironmentCategory{},
			},
		}

		models, diags := ConvertServiceTokenPermissionDataToModel(ctx, apiPermissions)

		require.False(t, diags.HasError())
		require.Len(t, models, 1)

		var writableEnvs []string
		diags = models[0].WritableEnvironmentCategories.ElementsAs(ctx, &writableEnvs, false)
		require.False(t, diags.HasError())

		assert.NotContains(
			t,
			writableEnvs,
			dbt_cloud.EnvironmentCategory_All,
			"an empty writable_environment_categories from the API must not be silently reported as all in state",
		)
		assert.Empty(t, writableEnvs)
	})
}
