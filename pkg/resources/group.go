package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,

		Schema: map[string]*schema.Schema{
			"is_active": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the group is active",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Group name",
			},
			"assign_by_default": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether or not to assign this group to users by default",
			},
			"sso_mapping_groups": &schema.Schema{
				Type:        schema.TypeList,
				Required:    true,
				Description: "SSO mapping group names for this group",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	assignByDefault := d.Get("assign_by_default").(bool)
	ssoMappingGroupsRaw := d.Get("sso_mapping_groups").([]interface{})
	ssoMappingGroups := make([]string, len(ssoMappingGroupsRaw))
	for i, _ := range ssoMappingGroupsRaw {
		ssoMappingGroups[i] = ssoMappingGroupsRaw[i].(string)
	}

	group, err := c.CreateGroup(name, assignByDefault, ssoMappingGroups)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(*group.ID))

	resourceGroupRead(ctx, d, m)

	return diags
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	groupID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

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

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fmt.Printf("Groups do not currently support updates")
	return resourceGroupRead(ctx, d, m)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	groupID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	group, err := c.GetGroup(groupID)
	if err != nil {
		return diag.FromErr(err)
	}

	group.State = dbt_cloud.STATE_DELETED
	_, err = c.UpdateGroup(groupID, *group)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
