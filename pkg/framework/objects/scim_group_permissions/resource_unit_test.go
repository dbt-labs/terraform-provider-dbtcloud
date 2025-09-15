package scim_group_permissions_test

import (
	"context"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/scim_group_permissions"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestScimGroupPermissionsResource_Schema(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schemaRequest := resource.SchemaRequest{}
	schemaResponse := &resource.SchemaResponse{}

	// Instantiate the resource and call Schema method
	r := scim_group_permissions.ScimGroupPermissionsResource()
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

func TestScimGroupPermissionsResource_Metadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadataRequest := resource.MetadataRequest{
		ProviderTypeName: "dbtcloud",
	}
	metadataResponse := &resource.MetadataResponse{}

	// Instantiate the resource and call Metadata method
	r := scim_group_permissions.ScimGroupPermissionsResource()
	r.Metadata(ctx, metadataRequest, metadataResponse)

	expectedTypeName := "dbtcloud_scim_group_permissions"
	if metadataResponse.TypeName != expectedTypeName {
		t.Errorf("Expected TypeName to be %s, got %s", expectedTypeName, metadataResponse.TypeName)
	}
}
