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
	licenseTypes = []string{
		"developer",
		"read_only",
		"it",
	}
)

func ResourceLicenseMap() *schema.Resource {
	return &schema.Resource{
		CreateContext: licenseMapCreate,
		ReadContext:   licenseMapRead,
		UpdateContext: licenseMapUpdate,
		DeleteContext: licenseMapDelete,

		Schema: map[string]*schema.Schema{
			"license_type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "License type",
				ValidateFunc: validation.StringInSlice(licenseTypes, false),
			},
			"sso_license_mapping_groups": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "SSO license mapping group names for this group",
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
	ssoLicenseMappingGroupsRaw := d.Get("sso_license_mapping_groups").([]interface{})
	ssoLicenseMappingGroups := make([]string, len(ssoLicenseMappingGroupsRaw))
	for i, _ := range ssoLicenseMappingGroupsRaw {
		ssoLicenseMappingGroups[i] = ssoLicenseMappingGroupsRaw[i].(string)
	}

	licenseMap, err := c.CreateLicenseMap(licenseType, ssoLicenseMappingGroups)
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
	if err := d.Set("sso_license_mapping_groups", licenseMap.SSOLicenseMappingGroups); err != nil {
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

	if d.HasChange("license_type") || d.HasChange("sso_license_mapping_groups") {
		licenseMap, err := c.GetLicenseMap(licenseMapID)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("license_type") {
			licenseType := d.Get("license_type").(string)
			licenseMap.LicenseType = licenseType
		}
		if d.HasChange("sso_license_mapping_groups") {
			ssoLicenseMappingGroups := d.Get("sso_license_mapping_groups").([]string)
			licenseMap.SSOLicenseMappingGroups = ssoLicenseMappingGroups
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
