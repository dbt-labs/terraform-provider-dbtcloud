package resources

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	permissionSets = []string{
		"owner",
		"member",
		"account_admin",
		"security_admin",
		"billing_admin",
		"admin",
		"database_admin",
		"git_admin",
		"team_admin",
		"job_admin",
		"job_viewer",
		"analyst",
		"developer",
		"stakeholder",
		"readonly",
		"project_creator",
		"account_viewer",
		"metadata_only",
		"semantic_layer_only",
		"webhooks_only",
	}
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
				Optional:    true,
				Default:     false,
				Description: "Whether or not to assign this group to users by default",
			},
			"sso_mapping_groups": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "SSO mapping group names for this group",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"group_permissions": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permission_set": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Set of permissions to apply",
							ValidateFunc: validation.StringInSlice(permissionSets, false),
						},
						"project_id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Project ID to apply this permission to for this group",
						},
						"all_projects": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether or not to apply this permission to all projects for this group",
						},
					},
				},
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceGroupCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
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

	groupPermissionsRaw := d.Get("group_permissions").(*schema.Set).List()
	groupPermissions := make([]dbt_cloud.GroupPermission, len(groupPermissionsRaw))
	for i, p := range groupPermissionsRaw {
		permission := p.(map[string]interface{})
		groupPermission := dbt_cloud.GroupPermission{
			GroupID:     *group.ID,
			AccountID:   group.AccountID,
			Set:         permission["permission_set"].(string),
			ProjectID:   permission["project_id"].(int),
			AllProjects: permission["all_projects"].(bool),
		}
		groupPermissions[i] = groupPermission
	}

	_, err = c.UpdateGroupPermissions(*group.ID, groupPermissions)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(*group.ID))

	resourceGroupRead(ctx, d, m)

	return diags
}

func resourceGroupRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	groupID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	group, err := c.GetGroup(groupID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
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
	permissions := make([]interface{}, len(group.Permissions))
	for i, permission := range group.Permissions {
		p := make(map[string]interface{})
		p["permission_set"] = permission.Set
		p["project_id"] = permission.ProjectID
		p["all_projects"] = permission.AllProjects
		permissions[i] = p
	}
	if err := d.Set("group_permissions", permissions); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceGroupUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	groupID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("name") ||
		d.HasChange("assign_by_default") ||
		d.HasChange("sso_mapping_groups") {
		group, err := c.GetGroup(groupID)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("name") {
			name := d.Get("name").(string)
			group.Name = name
		}
		if d.HasChange("assign_by_default") {
			assignByDefault := d.Get("assign_by_default").(bool)
			group.AssignByDefault = assignByDefault
		}
		if d.HasChange("sso_mapping_groups") {
			ssoMappingGroupsRaw := d.Get("sso_mapping_groups").([]interface{})
			ssoMappingGroups := make([]string, len(ssoMappingGroupsRaw))
			for i := range ssoMappingGroupsRaw {
				ssoMappingGroups[i] = ssoMappingGroupsRaw[i].(string)
			}
			group.SSOMappingGroups = ssoMappingGroups
		}
		_, err = c.UpdateGroup(groupID, *group)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("group_permissions") {
		// TODO(GtheSheep): Extract this to a function
		groupPermissionsRaw := d.Get("group_permissions").(*schema.Set).List()
		groupPermissions := make([]dbt_cloud.GroupPermission, len(groupPermissionsRaw))
		for i, p := range groupPermissionsRaw {
			permission := p.(map[string]interface{})
			groupPermission := dbt_cloud.GroupPermission{
				GroupID:     groupID,
				AccountID:   c.AccountID,
				Set:         permission["permission_set"].(string),
				ProjectID:   permission["project_id"].(int),
				AllProjects: permission["all_projects"].(bool),
			}
			groupPermissions[i] = groupPermission
		}
		_, err = c.UpdateGroupPermissions(groupID, groupPermissions)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceGroupRead(ctx, d, m)
}

func resourceGroupDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
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
