package scim_group_partial_permissions_test

import (
	"context"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/scim_group_partial_permissions"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
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

// TestUnionByDeduplicatesPermissions tests that UnionBy correctly deduplicates
// permissions when combining remote and new permissions during Create operations.
// This is a regression test for the permission duplication bug.
func TestUnionByDeduplicatesPermissions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Simulate remote permissions that already exist (e.g., from another resource or manual creation)
	remotePermissions := []scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		{
			PermissionSet: types.StringValue("account_admin"),
			ProjectID:     types.Int64Null(),
			AllProjects:   types.BoolValue(true),
			WritableEnvironmentCategories: types.SetNull(types.StringType),
		},
		{
			PermissionSet: types.StringValue("member"),
			ProjectID:     types.Int64Null(),
			AllProjects:   types.BoolValue(true),
			WritableEnvironmentCategories: types.SetNull(types.StringType),
		},
	}

	// Simulate new permissions that this resource wants to add
	// Note: member already exists in remote, but we're trying to add it again
	missingPermissions := []scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		{
			PermissionSet: types.StringValue("member"), // Duplicate!
			ProjectID:     types.Int64Null(),
			AllProjects:   types.BoolValue(true),
			WritableEnvironmentCategories: types.SetNull(types.StringType),
		},
		{
			PermissionSet: types.StringValue("developer"),
			ProjectID:     types.Int64Value(123),
			AllProjects:   types.BoolValue(false),
			WritableEnvironmentCategories: func() types.Set {
				set, _ := types.SetValueFrom(ctx, types.StringType, []string{"development"})
				return set
			}(),
		},
	}

	// Use UnionBy to combine (this is what the fixed Create function does)
	result := helper.UnionBy(
		remotePermissions,
		missingPermissions,
		scim_group_partial_permissions.CompareScimGroupPartialPermissions,
	)

	// Should have 3 unique permissions (account_admin, member, developer)
	// NOT 4 (which would indicate member was duplicated)
	if len(result) != 3 {
		t.Errorf("Expected 3 deduplicated permissions, got %d", len(result))
	}

	// Verify we have exactly one of each
	permissionSets := make(map[string]int)
	for _, perm := range result {
		permissionSets[perm.PermissionSet.ValueString()]++
	}

	expectedCounts := map[string]int{
		"account_admin": 1,
		"member":        1,
		"developer":     1,
	}

	for permSet, expectedCount := range expectedCounts {
		if count, ok := permissionSets[permSet]; !ok {
			t.Errorf("Missing permission_set %s", permSet)
		} else if count != expectedCount {
			t.Errorf("Permission %s appeared %d times, expected %d", permSet, count, expectedCount)
		}
	}
}

// TestUnionByHandlesConcurrentResourceCreation tests the scenario where
// multiple resources manage the same group and create near-simultaneously.
// This simulates the federated permission management pattern.
func TestUnionByHandlesConcurrentResourceCreation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Initial state: One pre-existing permission (e.g., manually added)
	initialRemote := []scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		{
			PermissionSet: types.StringValue("account_admin"),
			ProjectID:     types.Int64Null(),
			AllProjects:   types.BoolValue(true),
			WritableEnvironmentCategories: types.SetNull(types.StringType),
		},
	}

	// Resource A wants to add "member"
	resourceAMissing := []scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		{
			PermissionSet: types.StringValue("member"),
			ProjectID:     types.Int64Null(),
			AllProjects:   types.BoolValue(true),
			WritableEnvironmentCategories: types.SetNull(types.StringType),
		},
	}

	// Resource B wants to add "developer" (reads same initial state due to timing)
	resourceBMissing := []scim_group_partial_permissions.ScimGroupPartialPermissionModel{
		{
			PermissionSet: types.StringValue("developer"),
			ProjectID:     types.Int64Value(123),
			AllProjects:   types.BoolValue(false),
			WritableEnvironmentCategories: func() types.Set {
				set, _ := types.SetValueFrom(ctx, types.StringType, []string{"development"})
				return set
			}(),
		},
	}

	// Resource A combines with UnionBy
	resourceAResult := helper.UnionBy(
		initialRemote,
		resourceAMissing,
		scim_group_partial_permissions.CompareScimGroupPartialPermissions,
	)

	// Resource B combines with UnionBy (still sees initial remote)
	resourceBResult := helper.UnionBy(
		initialRemote,
		resourceBMissing,
		scim_group_partial_permissions.CompareScimGroupPartialPermissions,
	)

	// Both should have account_admin + their respective permission
	if len(resourceAResult) != 2 {
		t.Errorf("Resource A: expected 2 permissions, got %d", len(resourceAResult))
	}
	if len(resourceBResult) != 2 {
		t.Errorf("Resource B: expected 2 permissions, got %d", len(resourceBResult))
	}

	// Simulate what happens when Resource B runs after Resource A completes
	// Resource B should read the updated remote state and combine properly
	updatedRemote := resourceAResult // Now includes account_admin + member
	resourceBFinal := helper.UnionBy(
		updatedRemote,
		resourceBMissing,
		scim_group_partial_permissions.CompareScimGroupPartialPermissions,
	)

	// Should have all 3 permissions with no duplicates
	if len(resourceBFinal) != 3 {
		t.Errorf("Final state: expected 3 permissions, got %d", len(resourceBFinal))
	}

	permissionSets := make(map[string]int)
	for _, perm := range resourceBFinal {
		permissionSets[perm.PermissionSet.ValueString()]++
	}

	// Verify no duplicates
	for permSet, count := range permissionSets {
		if count != 1 {
			t.Errorf("Permission %s appeared %d times, expected 1 (duplicate detected!)", permSet, count)
		}
	}
}
