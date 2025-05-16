package data_sources

import (
	"context"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var extendedAttributesSchema = map[string]*schema.Schema{
	"extended_attributes_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID of the extended attributes",
	},
	"project_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID the extended attributes refers to",
	},
	"state": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "The state of the extended attributes (1 = active, 2 = inactive)",
	},
	"extended_attributes": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "A JSON string listing the extended attributes mapping",
	},
}

func DatasourceExtendedAttributes() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceExtendedAttributesRead,
		Schema:      extendedAttributesSchema,
	}
}

func datasourceExtendedAttributesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	extendedAttributesID := d.Get("extended_attributes_id").(int)
	projectID := d.Get("project_id").(int)

	extendedAttributes, err := c.GetExtendedAttributes(projectID, extendedAttributesID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("extended_attributes_id", extendedAttributes.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", extendedAttributes.State); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", extendedAttributes.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("extended_attributes", string(extendedAttributes.ExtendedAttributes)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(*extendedAttributes.ID))

	return diags
}
