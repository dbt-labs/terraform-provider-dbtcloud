package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceUserGroups() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserGroupsCreate,
		ReadContext:   resourceUserGroupsRead,
		UpdateContext: resourceUserGroupsUpdate,
		DeleteContext: resourceUserGroupsDelete,

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The internal ID of a dbt Cloud user",
			},
			"group_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "IDs of the groups to assign to the user. If additional groups were assigned manually in dbt Cloud, they will be removed.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: `
Assigns a set of dbt Cloud groups to a given User ID. 

~> If additional groups were assigned manually in dbt Cloud, they will be removed. The full list of groups need to be provided as config.
		
~> This resource does not currently support deletion (e.g. a deleted resource will stay as-is in dbt Cloud).
This is intentional in order to prevent accidental deletion of all users groups assigned to a user.
If you would like a different behavior, please open an issue on GitHub. To remove all groups for a user, set "group_ids" to the empty set "[]".
`,
	}
}

func checkGroupsAssigned(groupIDs []int, groupsAssigned *dbt_cloud.AssignUserGroupsResponse) error {
	groupIDsAssignedMap := map[int]bool{}
	for _, group := range groupsAssigned.Data {
		groupIDsAssignedMap[*group.ID] = true
	}

	for _, groupID := range groupIDs {
		if !groupIDsAssignedMap[groupID] {
			return fmt.Errorf("the Group %d was not assigned to the user (it's possible that it doesn't exist and needs to be removed from the config)", groupID)
		}
	}

	return nil
}

func resourceUserGroupsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	userID := d.Get("user_id").(int)
	groupIDsRaw := d.Get("group_ids").(*schema.Set)

	groupIDs := []int{}
	for _, groupID := range groupIDsRaw.List() {
		groupIDs = append(groupIDs, groupID.(int))
	}

	groupsAssigned, err := c.AssignUserGroups(userID, groupIDs)
	if err != nil {
		return diag.FromErr(err)
	}

	// dbt Cloud returns a 200 even if some groups don't exist. We need to check that all groups were assigned.
	if err := checkGroupsAssigned(groupIDs, groupsAssigned); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(userID))

	resourceUserGroupsRead(ctx, d, m)

	return diags
}

func resourceUserGroupsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	userID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	userGroups, err := c.GetUserGroups(userID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("user_id", userID); err != nil {
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

func resourceUserGroupsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	userID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("user_id") || d.HasChange("group_ids") {

		groupIDsRaw := d.Get("group_ids").(*schema.Set)
		groupIDs := []int{}
		for _, groupID := range groupIDsRaw.List() {
			groupIDs = append(groupIDs, groupID.(int))
		}

		groupsAssigned, err := c.AssignUserGroups(userID, groupIDs)
		if err != nil {
			return diag.FromErr(err)
		}

		// dbt Cloud returns a 200 even if some groups don't exist. We need to check that all groups were assigned.
		if err := checkGroupsAssigned(groupIDs, groupsAssigned); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceUserGroupsRead(ctx, d, m)
}

func resourceUserGroupsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// No-op implementation. Log a message or simply return.
	tflog.Warn(ctx, "[WARN] dbtcloud_user_groups does not support delete.")
	return nil
}
