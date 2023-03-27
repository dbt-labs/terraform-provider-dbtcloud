package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	authTypes = []string{
		"password",
		"keypair",
	}
)

func ResourcePostgresCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePostgresCredentialCreate,
		ReadContext:   resourcePostgresCredentialRead,
		UpdateContext: resourcePostgresCredentialUpdate,
		DeleteContext: resourcePostgresCredentialDelete,

		Schema: map[string]*schema.Schema{
			"is_active": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the Postgres/Redshift/AlloyDB credential is active",
			},
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Project ID to create the Postgres/Redshift/AlloyDB credential in",
			},
			"credential_id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The system Postgres/Redshift/AlloyDB credential ID",
			},
			"default_schema": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Default schema name",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username for Postgres/Redshift/AlloyDB",
			},
			"type": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Type of connection. One of (postgres/redshift/alloydb)",
			},
			"password": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				Description:   "Password for Postgres/Redshift/AlloyDB",
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

func resourcePostgresCredentialCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	authType := d.Get("auth_type").(string)
	default_schema := d.Get("default_schema").(string)
	type_ := d.Get("user").(string)
	username := d.Get("user").(string)
	password := d.Get("password").(string)
	numThreads := d.Get("num_threads").(int)

	postgresCredential, err := c.CreatePostgresCredential(projectId, isActive, type_, default_schema, username, password, privateKey, numThreads)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", postgresCredential.Project_Id, dbt_cloud.ID_DELIMITER, *postgresCredential.ID))

	resourcePostgresCredentialRead(ctx, d, m)

	return diags
}

func resourcePostgresCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	postgresCredentialId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	postgresCredential, err := c.GetPostgresCredential(projectId, postgresCredentialId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("credential_id", postgresCredentialId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_active", postgresCredential.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", postgresCredential.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_schema", postgresCredential.defaultSchema); err != nil {
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

func resourcePostgresCredentialUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	postgresCredentialId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("default_schema") || d.HasChange("username") || d.HasChange("password") || d.HasChange("num_threads") {
		postgresCredential, err := c.GetPostgresCredential(projectId, postgresCredentialId)
		if err != nil {
			return diag.FromErr(err)
		}
		if d.HasChange("default_schema") {
			schema := d.Get("default_schema").(string)
			postgresCredential.Schema = schema
		}
		if d.HasChange("username") {
			user := d.Get("username").(string)
			postgresCredential.User = user
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

func resourcePostgresCredentialDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	postgresCredentialIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	_, err := c.DeletePostgresCredential(postgresCredentialIdString, projectIdString)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
