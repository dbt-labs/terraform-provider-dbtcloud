package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var projectArtefactsSchema = map[string]*schema.Schema{
	"project_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID",
		ForceNew:    true,
	},
	"docs_job_id": {
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "Docs Job ID",
	},
	"freshness_job_id": {
		Type:        schema.TypeInt,
		Optional:    true,
		Description: "Freshness Job ID",
	},
}

func ResourceProjectArtefacts() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectArtefactsCreate,
		ReadContext:   resourceProjectArtefactsRead,
		UpdateContext: resourceProjectArtefactsUpdate,
		DeleteContext: resourceProjectArtefactsDelete,

		Schema:             projectArtefactsSchema,
		DeprecationMessage: "Please remove this deprecated resource from your project. It will be deleted from the provider in a future version. dbt Explorer replaces the need for this config.",
		Description:        "[Deprecated] Resource for mentioning what jobs are the source of truth for the legacy dbt Docs and dbt Source Freshness pages. dbt Explorer doesn't require this config anymore.",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceProjectArtefactsCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID := d.Get("project_id").(int)
	docsJobID := d.Get("docs_job_id").(int)
	freshnessJobID := d.Get("freshness_job_id").(int)
	projectIDString := strconv.Itoa(projectID)

	project, err := c.GetProject(projectIDString)
	if err != nil {
		return diag.FromErr(err)
	}

	if docsJobID != 0 {
		project.DocsJobId = &docsJobID
	} else {
		project.DocsJobId = nil
	}

	if freshnessJobID != 0 {
		project.FreshnessJobId = &freshnessJobID
	} else {
		project.FreshnessJobId = nil
	}

	_, err = c.UpdateProject(projectIDString, *project)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", *project.ID))

	resourceProjectArtefactsRead(ctx, d, m)

	return diags
}

func resourceProjectArtefactsRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	projectIDString := strconv.Itoa(projectID)

	project, err := c.GetProject(projectIDString)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("project_id", project.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("docs_job_id", project.DocsJobId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("freshness_job_id", project.FreshnessJobId); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceProjectArtefactsUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {

	c := m.(*dbt_cloud.Client)

	projectIDString := d.Id()

	if d.HasChange("docs_job_id") || d.HasChange("freshness_job_id") {
		project, err := c.GetProject(projectIDString)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("docs_job_id") {
			docsJobId := d.Get("docs_job_id").(int)
			if docsJobId != 0 {
				project.DocsJobId = &docsJobId
			} else {
				project.DocsJobId = nil
			}
		}

		if d.HasChange("freshness_job_id") {
			freshnessJobId := d.Get("freshness_job_id").(int)
			if freshnessJobId != 0 {
				project.FreshnessJobId = &freshnessJobId
			} else {
				project.FreshnessJobId = nil
			}
		}

		_, err = c.UpdateProject(projectIDString, *project)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceProjectArtefactsRead(ctx, d, m)
}

func resourceProjectArtefactsDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectID := d.Get("project_id").(int)
	projectIDString := strconv.Itoa(projectID)

	project, err := c.GetProject(projectIDString)
	if err != nil {
		return diag.FromErr(err)
	}

	project.FreshnessJobId = nil
	project.DocsJobId = nil

	_, err = c.UpdateProject(projectIDString, *project)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
