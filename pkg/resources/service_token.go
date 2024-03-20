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

func ResourceServiceToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceTokenCreate,
		ReadContext:   resourceServiceTokenRead,
		UpdateContext: resourceServiceTokenUpdate,
		DeleteContext: resourceServiceTokenDelete,

		Schema: map[string]*schema.Schema{
			"uid": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service token UID (part of the token)",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service token name",
				ForceNew:    true,
			},
			"token_string": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Service token secret value (only accessible on creation))",
			},
			"state": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Service token state (1 is active, 2 is inactive)",
			},
			"service_token_permissions": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Permissions set for the service token",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permission_set": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Set of permissions to apply",
							ValidateFunc: validation.StringInSlice(
								dbt_cloud.PermissionSets,
								false,
							),
						},
						"project_id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Project ID to apply this permission to for this service token",
						},
						"all_projects": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether or not to apply this permission to all projects for this service token",
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

func resourceServiceTokenCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	state := d.Get("state").(int)

	serviceToken, err := c.CreateServiceToken(name, state)
	if err != nil {
		return diag.FromErr(err)
	}

	serviceTokenPermissionsRaw := d.Get("service_token_permissions").(*schema.Set).List()
	serviceTokenPermissions := make(
		[]dbt_cloud.ServiceTokenPermission,
		len(serviceTokenPermissionsRaw),
	)
	for i, p := range serviceTokenPermissionsRaw {
		permission := p.(map[string]interface{})
		serviceTokenPermission := dbt_cloud.ServiceTokenPermission{
			ServiceTokenID: *serviceToken.ID,
			AccountID:      serviceToken.AccountID,
			Set:            permission["permission_set"].(string),
			ProjectID:      permission["project_id"].(int),
			AllProjects:    permission["all_projects"].(bool),
		}
		serviceTokenPermissions[i] = serviceTokenPermission
	}

	_, err = c.UpdateServiceTokenPermissions(*serviceToken.ID, serviceTokenPermissions)
	if err != nil {
		return diag.FromErr(err)
	}

	// The string is only available the first time we ge create the token
	d.SetId(strconv.Itoa(*serviceToken.ID))
	if err := d.Set("token_string", serviceToken.TokenString); err != nil {
		return diag.FromErr(err)
	}
	resourceServiceTokenRead(ctx, d, m)

	return diags
}

func resourceServiceTokenRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	serviceTokenID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	serviceToken, err := c.GetServiceToken(serviceTokenID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("uid", serviceToken.UID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", serviceToken.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", serviceToken.State); err != nil {
		return diag.FromErr(err)
	}

	serviceTokenPermissions, err := c.GetServiceTokenPermissions(serviceTokenID)
	if err != nil {
		return diag.FromErr(err)
	}

	permissions := make([]interface{}, len(*serviceTokenPermissions))
	for i, permission := range *serviceTokenPermissions {
		p := make(map[string]interface{})
		p["permission_set"] = permission.Set
		p["project_id"] = permission.ProjectID
		p["all_projects"] = permission.AllProjects
		permissions[i] = p
	}
	if err := d.Set("service_token_permissions", permissions); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceServiceTokenUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	serviceTokenID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("name") || d.HasChange("state") {

		serviceToken, err := c.GetServiceToken(serviceTokenID)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("name") {
			name := d.Get("name").(string)
			serviceToken.Name = name
		}
		if d.HasChange("uid") {
			uid := d.Get("uid").(string)
			serviceToken.UID = uid
		}
		if d.HasChange("state") {
			state := d.Get("state").(int)
			serviceToken.State = state
		}

		_, err = c.UpdateServiceToken(serviceTokenID, *serviceToken)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("service_token_permissions") {
		serviceTokenPermissionsRaw := d.Get("service_token_permissions").(*schema.Set).List()
		serviceTokenPermissions := make(
			[]dbt_cloud.ServiceTokenPermission,
			len(serviceTokenPermissionsRaw),
		)
		for i, p := range serviceTokenPermissionsRaw {
			permission := p.(map[string]interface{})
			serviceTokenPermission := dbt_cloud.ServiceTokenPermission{
				ServiceTokenID: serviceTokenID,
				AccountID:      c.AccountID,
				Set:            permission["permission_set"].(string),
				ProjectID:      permission["project_id"].(int),
				AllProjects:    permission["all_projects"].(bool),
			}
			serviceTokenPermissions[i] = serviceTokenPermission
		}
		_, err = c.UpdateServiceTokenPermissions(serviceTokenID, serviceTokenPermissions)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceServiceTokenRead(ctx, d, m)
}

func resourceServiceTokenDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	serviceTokenID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.DeleteServiceToken(serviceTokenID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
