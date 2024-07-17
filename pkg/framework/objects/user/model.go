package user

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userDataSourceModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Email types.String `tfsdk:"email"`
}

type usersDataSourceModel struct {
	Users []userDataSourceModel `tfsdk:"users"`
}
