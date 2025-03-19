package user_groups

import "github.com/hashicorp/terraform-plugin-framework/types"

type UserGroupsResourceModel struct {
	ID types.String `tfsdk:"id"`
	UserID types.Int64 `tfsdk:"user_id"`
	GroupIDs types.Set `tfsdk:"group_ids"` 
}

type UserGroupsDataSourceModel struct {
	ID types.String `tfsdk:"id"`
	UserID types.Int64 `tfsdk:"user_id"`
	GroupIDs types.Set `tfsdk:"group_ids"`
}