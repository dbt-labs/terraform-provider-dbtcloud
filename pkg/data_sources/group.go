package data_sources

import (
	"context"
	"strconv"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var groupSchema = map[string]*schema.Schema{
	"group_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID of the group",
	},
	"is_active": &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the group is active",
	},
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Group name",
	},
	"assign_by_default": &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether or not to assign this group to users by default",
	},
	"sso_mapping_groups": &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "SSO mapping group names for this group",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
}

func DatasourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceGroupRead,
		Schema:      groupSchema,
	}
}

func datasourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	groupID := d.Get("group_id").(int)

	group, err := c.GetGroup(groupID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_active", group.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", group.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("assign_by_default", group.AssignByDefault); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sso_mapping_groups", group.SSOMappingGroups); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(*group.ID))

	return diags
}
