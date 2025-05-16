package data_sources

import (
	"context"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var userGroupsSchema = map[string]*schema.Schema{
	"user_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID of the user",
	},
	"group_ids": {
		Type:        schema.TypeSet,
		Computed:    true,
		Description: "IDs of the groups assigned to the user",
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
	},
}

func DatasourceUserGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceUserGroupsRead,
		Schema:      userGroupsSchema,
	}
}

func datasourceUserGroupsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	userID := d.Get("user_id").(int)

	userGroups, err := c.GetUserGroups(userID)
	if err != nil {
		return diag.FromErr(err)
	}

	groupIDs := []int{}
	for _, group := range userGroups.Groups {
		groupIDs = append(groupIDs, *group.ID)
	}
	if err := d.Set("group_ids", groupIDs); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(userID))

	return diags
}
