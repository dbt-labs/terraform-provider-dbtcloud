package user_groups

import "github.com/hashicorp/terraform-plugin-framework/types"

type ModelUserGroupsResourceModel struct {
	ID int `tfsdk:"id"`
	UserID types.Int64 `tfsdk:"user_id"`
	GroupIDs types.Set `tfsdk:"group_ids"` // define this in the schema.go
}