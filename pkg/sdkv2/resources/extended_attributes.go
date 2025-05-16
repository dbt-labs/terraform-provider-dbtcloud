package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceExtendedAttributes() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExtendedAttributesCreate,
		ReadContext:   resourceExtendedAttributesRead,
		UpdateContext: resourceExtendedAttributesUpdate,
		DeleteContext: resourceExtendedAttributesDelete,

		Schema: map[string]*schema.Schema{
			"extended_attributes_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Extended Attributes ID",
			},
			"state": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Extended Attributes state (1 is active, 2 is inactive)",
				Deprecated:  "Remove this attribute's configuration as it's no longer in use and the attribute will be removed in the next major version of the provider.",
			},
			"project_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Project ID to create the extended attributes in",
			},
			"extended_attributes": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A JSON string listing the extended attributes mapping. The keys are the connections attributes available in the `profiles.yml` for a given adapter. Any fields entered will override connection details or credentials set on the environment or project. To avoid incorrect Terraform diffs, it is recommended to create this string using `jsonencode` in your Terraform code. (see example)",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					var oldJSON, newJSON map[string]interface{}
					if err := json.Unmarshal([]byte(old), &oldJSON); err != nil {
						return false
					}
					if err := json.Unmarshal([]byte(new), &newJSON); err != nil {
						return false
					}
					return reflect.DeepEqual(oldJSON, newJSON)
				},
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "This resource allows setting extended attributes which can be assigned to a given environment ([see docs](https://docs.getdbt.com/docs/dbt-cloud-environments#extended-attributes)).<br/><br/>In dbt Cloud those values are provided as YML but in the provider they need to be provided as JSON (see example below).",
	}
}

func resourceExtendedAttributesCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	state := d.Get("state").(int)
	projectId := d.Get("project_id").(int)
	extendedAttributesValue := d.Get("extended_attributes").(string)
	extendedAttributesRaw := json.RawMessage([]byte(extendedAttributesValue))

	extendedAttributes, err := c.CreateExtendedAttributes(state, projectId, extendedAttributesRaw)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(
		fmt.Sprintf(
			"%d%s%d",
			extendedAttributes.ProjectID,
			dbt_cloud.ID_DELIMITER,
			*extendedAttributes.ID,
		),
	)

	resourceExtendedAttributesRead(ctx, d, m)

	return diags
}

func resourceExtendedAttributesRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics
	projectID, extendedAttributesID, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_extended_attributes",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	extendedAttributes, err := c.GetExtendedAttributes(projectID, extendedAttributesID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("extended_attributes_id", &extendedAttributes.ID); err != nil {
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

	return diags
}

// func changedExtendedAttributes(old, new interface{}) bool {

// 	oldStr := old.(string)
// 	newStr := new.(string)

// 	var objOld, objNew map[string]interface{}

// 	if err := json.Unmarshal([]byte(oldStr), &objOld); err != nil {
// 		panic(err)
// 	}
// 	if err := json.Unmarshal([]byte(newStr), &objNew); err != nil {
// 		panic(err)
// 	}

// 	return !reflect.DeepEqual(objOld, objNew)
// }

func resourceExtendedAttributesUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, extendedAttributesId, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_extended_attributes",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	// old, new := d.GetChange("extended_attributes")

	if d.HasChange("state") ||
		d.HasChange("project_id") ||
		d.HasChange("extended_attributes") {

		extendedAttributes, err := c.GetExtendedAttributes(projectId, extendedAttributesId)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("state") {
			state := d.Get("state").(int)
			extendedAttributes.State = state
		}
		if d.HasChange("project_id") {
			projectID := d.Get("project_id").(int)
			extendedAttributes.ProjectID = projectID
		}
		if d.HasChange("extended_attributes") {
			extendedAttributesValue := d.Get("extended_attributes").(string)
			extendedAttributes.ExtendedAttributes = json.RawMessage([]byte(extendedAttributesValue))
		}

		_, err = c.UpdateExtendedAttributes(projectId, extendedAttributesId, *extendedAttributes)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceExtendedAttributesRead(ctx, d, m)
}

func resourceExtendedAttributesDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectId, extendedAttributesId, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_extended_attributes",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.DeleteExtendedAttributes(projectId, extendedAttributesId)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
