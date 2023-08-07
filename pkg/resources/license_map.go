package resources

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceLicenseMap() *schema.Resource {
	return &schema.Resource{
		CreateContext: licenseMapCreate,
		ReadContext:   licenseMapRead,
		UpdateContext: licenseMapUpdate,
		DeleteContext: licenseMapDelete,

		Schema: map[string]*schema.Schema{
			"license_type": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "License type",
			},
			"sso_mapping_groups": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
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

func licenseMapCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	licenseType := d.Get("license_type").(string)
	ssoMappingGroupsRaw := d.Get("sso_mapping_groups").([]interface{})
	ssoMappingGroups := make([]string, len(ssoMappingGroupsRaw))
	for i, _ := range ssoMappingGroupsRaw {
		ssoMappingGroups[i] = ssoMappingGroupsRaw[i].(string)
	}

	licenseMap, err := c.CreateLicenseMap(licenseType, ssoMappingGroups)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(*licenseMap.ID))

	licenseMapRead(ctx, d, m)

	return diags
}

func licenseMapRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	licenseMapID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	licenseMap, err := c.GetLicenseMap(licenseMapID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("license_type", licenseMap.LicenseType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sso_mapping_groups", licenseMap.SSOMappingGroups); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func licenseMapUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	licenseMapID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("license_type") || d.HasChange("sso_mapping_groups") {
		licenseMap, err := c.GetLicenseMap(licenseMapID)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("license_type") {
			licenseType := d.Get("license_type").(string)
			licenseMap.LicenseType = licenseType
		}
		if d.HasChange("sso_mapping_groups") {
			ssoMappingGroups := d.Get("sso_mapping_groups").([]string)
			licenseMap.SSOMappingGroups = ssoMappingGroups
		}
		_, err = c.UpdateLicenseMap(licenseMapID, *licenseMap)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func licenseMapDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	licenseMapID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.DestroyLicenseMap(licenseMapID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
