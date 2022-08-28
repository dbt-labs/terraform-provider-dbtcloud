package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceDatabricksCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDatabricksCredentialCreate,
		ReadContext:   resourceDatabricksCredentialRead,
		UpdateContext: resourceDatabricksCredentialUpdate,
		DeleteContext: resourceDatabricksCredentialDelete,

		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Project ID to create the Databricks credential in",
			},
			"adapter_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Databricks adapter ID for the credential",
			},
			"credential_id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The system Databricks credential ID",
			},
			"target_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Target name",
			},
			"token": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Token for Databricks user",
			},
			"num_threads": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Number of threads to use",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceDatabricksCredentialCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId := d.Get("project_id").(int)
	adapterId := d.Get("adapter_id").(int)
	targetName := d.Get("target_name").(string)
	token := d.Get("token").(string)
	numThreads := d.Get("num_threads").(int)

	databricksCredential, err := c.CreateDatabricksCredential(projectId, "adapter", targetName, adapterId, numThreads, token)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", databricksCredential.Project_Id, dbt_cloud.ID_DELIMITER, *databricksCredential.ID))

	resourceDatabricksCredentialRead(ctx, d, m)

	return diags
}

func resourceDatabricksCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	databricksCredentialId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	databricksCredential, err := c.GetDatabricksCredential(projectId, databricksCredentialId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("credential_id", databricksCredentialId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", databricksCredential.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("adapter_id", databricksCredential.Adapter_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("target_name", databricksCredential.Target_Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("num_threads", databricksCredential.Threads); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("token", d.Get("token").(string)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDatabricksCredentialUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	databricksCredentialId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("num_threads") || d.HasChange("token") {
		databricksCredential, err := c.GetDatabricksCredential(projectId, databricksCredentialId)
		if err != nil {
			return diag.FromErr(err)
		}
		if d.HasChange("num_threads") {
			numThreads := d.Get("num_threads").(int)
			databricksCredential.Threads = numThreads
		}
		if d.HasChange("token") {
			token := d.Get("token").(string)
			databricksCredential.Credential_Details.Fields.Token.Value = token
		}

		_, err = c.UpdateDatabricksCredential(projectId, databricksCredentialId, *databricksCredential)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDatabricksCredentialRead(ctx, d, m)
}

func resourceDatabricksCredentialDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}
	databricksCredentialId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	databricksCredential, err := c.GetDatabricksCredential(projectId, databricksCredentialId)
	if err != nil {
		return diag.FromErr(err)
	}

	databricksCredential.State = dbt_cloud.STATE_DELETED
	_, err = c.UpdateDatabricksCredential(projectId, databricksCredentialId, *databricksCredential)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
