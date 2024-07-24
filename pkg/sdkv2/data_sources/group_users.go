package data_sources

import (
	"context"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var groupUsersSchema = map[string]*schema.Schema{
	"group_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID of the group",
	},
	"users": {
		Type:        schema.TypeSet,
		Computed:    true,
		Description: "List of users (map of ID and email) in the group",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"email": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	},
}

func DatasourceGroupUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceGroupUsersRead,
		Schema:      groupUsersSchema,
		Description: "Returns a list of users assigned to a specific dbt Cloud group",
	}
}

func datasourceGroupUsersRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	groupID := d.Get("group_id").(int)

	users, err := c.GetUsers()
	if err != nil {
		return diag.FromErr(err)
	}

	usersParams := []map[string]interface{}{}
	for _, user := range users {
		userGroups := user.Permissions[0].Groups

		userInGroup := false
		for _, userGroup := range userGroups {
			if userGroup.ID == groupID {
				userInGroup = true
				// we can stop looping
				break
			}
		}

		if userInGroup {
			userToAdd := map[string]interface{}{}

			userToAdd["id"] = user.ID
			userToAdd["email"] = user.Email
			usersParams = append(usersParams, userToAdd)
		}
	}
	if err := d.Set("users", usersParams); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(groupID))

	return diags
}
