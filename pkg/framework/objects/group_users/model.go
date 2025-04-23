package group_users

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GroupUsersDataSourceModel is the model for the resource
type GroupUsersDataSourceModel struct {
	ID      types.String          `tfsdk:"id"`
	GroupID types.Int64           `tfsdk:"group_id"`
	Users   []userDataSourceModel `tfsdk:"users"`
}

type userDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
}
