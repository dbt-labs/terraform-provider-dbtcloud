package scim_group_partial_permissions_test

import (
	"context"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/scim_group_partial_permissions"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestScimGroupPartialPermissionsResource_Schema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	// Instantiate the resource and call Schema method
	r := scim_group_partial_permissions.ScimGroupPartialPermissionsResource()
	r.Schema(ctx, schemaRequest, schemaResponse)

	// Verify that the schema was created without errors
	if schemaResponse.Diagnostics.HasError() {
		t.Fatalf("Schema method returned errors: %v", schemaResponse.Diagnostics.Errors())
	}

	// Verify that key attributes exist
	schema := schemaResponse.Schema
	if _, ok := schema.Attributes["group_id"]; !ok {
		t.Error("Schema missing group_id attribute")
	}
	if _, ok := schema.Attributes["permissions"]; !ok {
		t.Error("Schema missing permissions attribute")
	}
	if _, ok := schema.Attributes["id"]; !ok {
		t.Error("Schema missing id attribute")
	}
}

func TestScimGroupPartialPermissionsResource_Metadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadataRequest := resource.MetadataRequest{
		ProviderTypeName: "dbtcloud",
	}
	metadataResponse := &resource.MetadataResponse{}

	// Instantiate the resource and call Metadata method
	r := scim_group_partial_permissions.ScimGroupPartialPermissionsResource()
	r.Metadata(ctx, metadataRequest, metadataResponse)

	expectedTypeName := "dbtcloud_scim_group_partial_permissions"
	if metadataResponse.TypeName != expectedTypeName {
		t.Errorf("Expected TypeName to be %s, got %s", expectedTypeName, metadataResponse.TypeName)
	}
}

func TestCompareScimGroupPartialPermissions_NormalizesEmptyAndNull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test 1: Null vs Null should be equal
	perm1 := scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		PermissionSet:                 types.StringValue("job_admin"),
		ProjectID:                     types.Int64Null(),
		AllProjects:                   types.BoolValue(true),
		WritableEnvironmentCategories: types.SetNull(types.StringType),
	}
	perm2 := scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		PermissionSet:                 types.StringValue("job_admin"),
		ProjectID:                     types.Int64Null(),
		AllProjects:                   types.BoolValue(true),
		WritableEnvironmentCategories: types.SetNull(types.StringType),
	}
	if !scim_group_partial_permissions.CompareScimGroupPartialPermissions(perm1, perm2) {
		t.Error("Null vs Null should be equal")
	}

	// Test 2: Null vs Empty Set should be equal
	emptySet, _ := types.SetValueFrom(ctx, types.StringType, []string{})
	perm3 := scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		PermissionSet:                 types.StringValue("job_admin"),
		ProjectID:                     types.Int64Null(),
		AllProjects:                   types.BoolValue(true),
		WritableEnvironmentCategories: types.SetNull(types.StringType),
	}
	perm4 := scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		PermissionSet:                 types.StringValue("job_admin"),
		ProjectID:                     types.Int64Null(),
		AllProjects:                   types.BoolValue(true),
		WritableEnvironmentCategories: emptySet,
	}
	if !scim_group_partial_permissions.CompareScimGroupPartialPermissions(perm3, perm4) {
		t.Error("Null vs Empty Set should be equal")
	}

	// Test 3: Empty Set vs Empty Set should be equal
	emptySet2, _ := types.SetValueFrom(ctx, types.StringType, []string{})
	perm5 := scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		PermissionSet:                 types.StringValue("account_admin"),
		ProjectID:                     types.Int64Null(),
		AllProjects:                   types.BoolValue(true),
		WritableEnvironmentCategories: emptySet,
	}
	perm6 := scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		PermissionSet:                 types.StringValue("account_admin"),
		ProjectID:                     types.Int64Null(),
		AllProjects:                   types.BoolValue(true),
		WritableEnvironmentCategories: emptySet2,
	}
	if !scim_group_partial_permissions.CompareScimGroupPartialPermissions(perm5, perm6) {
		t.Error("Empty Set vs Empty Set should be equal")
	}

	// Test 4: Non-empty set should NOT equal null
	nonEmptySet, _ := types.SetValueFrom(ctx, types.StringType, []string{"development", "staging"})
	perm7 := scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		PermissionSet:                 types.StringValue("developer"),
		ProjectID:                     types.Int64Value(123),
		AllProjects:                   types.BoolValue(false),
		WritableEnvironmentCategories: types.SetNull(types.StringType),
	}
	perm8 := scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		PermissionSet:                 types.StringValue("developer"),
		ProjectID:                     types.Int64Value(123),
		AllProjects:                   types.BoolValue(false),
		WritableEnvironmentCategories: nonEmptySet,
	}
	if scim_group_partial_permissions.CompareScimGroupPartialPermissions(perm7, perm8) {
		t.Error("Non-empty set should NOT equal null")
	}

	// Test 5: Matching non-empty sets should be equal
	nonEmptySet2, _ := types.SetValueFrom(ctx, types.StringType, []string{"development", "staging"})
	perm9 := scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		PermissionSet:                 types.StringValue("developer"),
		ProjectID:                     types.Int64Value(123),
		AllProjects:                   types.BoolValue(false),
		WritableEnvironmentCategories: nonEmptySet,
	}
	perm10 := scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		PermissionSet:                 types.StringValue("developer"),
		ProjectID:                     types.Int64Value(123),
		AllProjects:                   types.BoolValue(false),
		WritableEnvironmentCategories: nonEmptySet2,
	}
	if !scim_group_partial_permissions.CompareScimGroupPartialPermissions(perm9, perm10) {
		t.Error("Matching non-empty sets should be equal")
	}
}

func TestConvertScimGroupPartialPermissionDataToModel_NormalizesEmptyArray(t *testing.T) {
	t.Parallel()

	// Test that API response with empty array gets normalized to null in Terraform
	apiPermissions := []dbt_cloud.GroupPermission{
		{
			AccountID:                     123,
			GroupID:                       456,
			Set:                           "job_admin",
			AllProjects:                   true,
			ProjectID:                     0,
			WritableEnvironmentCategories: []string{}, // Empty array from API
		},
		{
			AccountID:                     123,
			GroupID:                       456,
			Set:                           "developer",
			AllProjects:                   false,
			ProjectID:                     789,
			WritableEnvironmentCategories: []string{"development", "staging"}, // Non-empty array
		},
	}

	models := scim_group_partial_permissions.ConvertScimGroupPartialPermissionDataToModel(apiPermissions)

	if len(models) != 2 {
		t.Fatalf("Expected 2 models, got %d", len(models))
	}

	// First permission should have null writable_environment_categories (normalized from empty array)
	if !models[0].WritableEnvironmentCategories.IsNull() {
		t.Error("Expected writable_environment_categories to be null for job_admin (was empty array in API)")
	}

	// Second permission should have non-null writable_environment_categories
	if models[1].WritableEnvironmentCategories.IsNull() {
		t.Error("Expected writable_environment_categories to be non-null for developer")
	}
	if len(models[1].WritableEnvironmentCategories.Elements()) != 2 {
		t.Errorf("Expected 2 elements in writable_environment_categories, got %d", len(models[1].WritableEnvironmentCategories.Elements()))
	}
}

func TestConvertScimGroupPartialPermissionModelToData_SendsEmptyArray(t *testing.T) {
	t.Parallel()

	// Test that null in Terraform gets sent as empty array to API
	models := []scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		{
			PermissionSet:                 types.StringValue("account_admin"),
			ProjectID:                     types.Int64Null(),
			AllProjects:                   types.BoolValue(true),
			WritableEnvironmentCategories: types.SetNull(types.StringType), // Null in Terraform
		},
	}

	apiPermissions := scim_group_partial_permissions.ConvertScimGroupPartialPermissionModelToData(models, 456, 123)

	if len(apiPermissions) != 1 {
		t.Fatalf("Expected 1 permission, got %d", len(apiPermissions))
	}

	// Should send empty array (not nil) to API
	if apiPermissions[0].WritableEnvironmentCategories == nil {
		t.Error("Expected writable_environment_categories to be empty array, got nil")
	}
	if len(apiPermissions[0].WritableEnvironmentCategories) != 0 {
		t.Errorf("Expected writable_environment_categories to be empty, got %d elements", len(apiPermissions[0].WritableEnvironmentCategories))
	}
}
