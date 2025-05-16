package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	warehouseTypes = []string{
		"postgres",
		"redshift",
	}
)

func ResourcePostgresCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePostgresCredentialCreate,
		ReadContext:   resourcePostgresCredentialRead,
		UpdateContext: resourcePostgresCredentialUpdate,
		DeleteContext: resourcePostgresCredentialDelete,

		Schema: map[string]*schema.Schema{
			"is_active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the Postgres/Redshift/AlloyDB credential is active",
			},
			"project_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID to create the Postgres/Redshift/AlloyDB credential in",
			},
			"credential_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The system Postgres/Redshift/AlloyDB credential ID",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of connection. One of (postgres/redshift). Use postgres for alloydb connections",
				ValidateFunc: validation.StringInSlice(warehouseTypes, false),
			},
			"default_schema": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Default schema name",
			},
			"target_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
				Description: "Default schema name",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username for Postgres/Redshift/AlloyDB",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password for Postgres/Redshift/AlloyDB",
			},
			"num_threads": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of threads to use",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourcePostgresCredentialCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	type_ := d.Get("type").(string)
	defaultSchema := d.Get("default_schema").(string)
	targetName := d.Get("target_name").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	numThreads := d.Get("num_threads").(int)

	postgresCredential, err := c.CreatePostgresCredential(
		projectId,
		isActive,
		type_,
		defaultSchema,
		targetName,
		username,
		password,
		numThreads,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(
		fmt.Sprintf(
			"%d%s%d",
			postgresCredential.Project_Id,
			dbt_cloud.ID_DELIMITER,
			*postgresCredential.ID,
		),
	)

	resourcePostgresCredentialRead(ctx, d, m)

	return diags
}

func resourcePostgresCredentialRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId, postgresCredentialId, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_postgres_credential",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	postgresCredential, err := c.GetPostgresCredential(projectId, postgresCredentialId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	postgresCredential.Password = d.Get("password").(string)

	if err := d.Set("credential_id", postgresCredentialId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_active", postgresCredential.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", postgresCredential.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", postgresCredential.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_schema", postgresCredential.Default_Schema); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("target_name", postgresCredential.Target_Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", postgresCredential.Username); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password", postgresCredential.Password); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("num_threads", postgresCredential.Threads); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourcePostgresCredentialUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, postgresCredentialId, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_postgres_credential",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("type") || d.HasChange("default_schema") || d.HasChange("target_name") ||
		d.HasChange("username") ||
		d.HasChange("password") ||
		d.HasChange("num_threads") {
		postgresCredential, err := c.GetPostgresCredential(projectId, postgresCredentialId)
		if err != nil {
			return diag.FromErr(err)
		}
		if d.HasChange("type") {
			default_schema := d.Get("type").(string)
			postgresCredential.Type = default_schema
		}
		if d.HasChange("default_schema") {
			default_schema := d.Get("default_schema").(string)
			postgresCredential.Default_Schema = default_schema
		}
		if d.HasChange("target_name") {
			default_schema := d.Get("target_name").(string)
			postgresCredential.Target_Name = default_schema
		}
		if d.HasChange("username") {
			username := d.Get("username").(string)
			postgresCredential.Username = username
		}
		if d.HasChange("password") {
			password := d.Get("password").(string)
			postgresCredential.Password = password
		}
		if d.HasChange("num_threads") {
			numThreads := d.Get("num_threads").(int)
			postgresCredential.Threads = numThreads
		}

		_, err = c.UpdatePostgresCredential(projectId, postgresCredentialId, *postgresCredential)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourcePostgresCredentialRead(ctx, d, m)
}

func resourcePostgresCredentialDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString, postgresCredentialIdString, err := helper.SplitIDToStrings(
		d.Id(),
		"dbtcloud_postgres_credential",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.DeleteCredential(postgresCredentialIdString, projectIdString)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
