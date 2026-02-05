package profile

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProfileResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ProfileID            types.Int64  `tfsdk:"profile_id"`
	ProjectID            types.Int64  `tfsdk:"project_id"`
	Key                  types.String `tfsdk:"key"`
	ConnectionID         types.Int64  `tfsdk:"connection_id"`
	CredentialsID        types.Int64  `tfsdk:"credentials_id"`
	ExtendedAttributesID types.Int64  `tfsdk:"extended_attributes_id"`
}

type ProfileDataSourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ProfileID            types.Int64  `tfsdk:"profile_id"`
	ProjectID            types.Int64  `tfsdk:"project_id"`
	Key                  types.String `tfsdk:"key"`
	ConnectionID         types.Int64  `tfsdk:"connection_id"`
	CredentialsID        types.Int64  `tfsdk:"credentials_id"`
	ExtendedAttributesID types.Int64  `tfsdk:"extended_attributes_id"`
}

type ProfilesDataSourceModel struct {
	ProjectID types.Int64              `tfsdk:"project_id"`
	Profiles  []ProfileDataSourceModel `tfsdk:"profiles"`
}
