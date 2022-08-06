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

func ResourceSnowflakeCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSnowflakeCredentialCreate,
		ReadContext:   resourceSnowflakeCredentialRead,
		UpdateContext: resourceSnowflakeCredentialUpdate,
		DeleteContext: resourceSnowflakeCredentialDelete,

		Schema: map[string]*schema.Schema{
			"is_active": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the Snowflake credential is active",
			},
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Project ID to create the Snowflake credential in",
			},
			"credential_id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The system Snowflake credential ID",
			},
			"auth_type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of Snowflake credential ('password' or 'keypair')",
				ValidateFunc: validation.StringInSlice(authTypes, false),
			},
			"schema": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Default schema name",
			},
			"user": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username for Snowflake",
			},
			"password": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				Description:   "Password for Snowflake",
				ConflictsWith: []string{"private_key", "private_key_passphrase"},
			},
			"private_key": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				Description:   "Private key for Snowflake",
				ConflictsWith: []string{"password"},
			},
			"private_key_passphrase": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				Description:   "Private key passphrase for Snowflake",
				ConflictsWith: []string{"password"},
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

func resourceSnowflakeCredentialCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	authType := d.Get("auth_type").(string)
	schema := d.Get("schema").(string)
	user := d.Get("user").(string)
	password := d.Get("password").(string)
	privateKey := d.Get("private_key").(string)
	privateKeyPassphrase := d.Get("private_key_passphrase").(string)
	numThreads := d.Get("num_threads").(int)

	snowflakeCredential, err := c.CreateSnowflakeCredential(projectId, "snowflake", isActive, schema, user, password, privateKey, privateKeyPassphrase, authType, numThreads)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", snowflakeCredential.Project_Id, dbt_cloud.ID_DELIMITER, *snowflakeCredential.ID))

	resourceSnowflakeCredentialRead(ctx, d, m)

	return diags
}

func resourceSnowflakeCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	snowflakeCredentialId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	snowflakeCredential, err := c.GetSnowflakeCredential(projectId, snowflakeCredentialId)
	if err != nil {
		return diag.FromErr(err)
	}
	// TODO: Manage passwords better
	if snowflakeCredential.Auth_Type == "password" {
		snowflakeCredential.Password = d.Get("password").(string)
	}
	if snowflakeCredential.Auth_Type == "keypair" {
		snowflakeCredential.PrivateKey = d.Get("private_key").(string)
		snowflakeCredential.PrivateKeyPassphrase = d.Get("private_key_passphrase").(string)
	}

	if err := d.Set("credential_id", snowflakeCredentialId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_active", snowflakeCredential.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", snowflakeCredential.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("auth_type", snowflakeCredential.Auth_Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schema", snowflakeCredential.Schema); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user", snowflakeCredential.User); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("password", snowflakeCredential.Password); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("private_key", snowflakeCredential.PrivateKey); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("private_key_passphrase", snowflakeCredential.PrivateKeyPassphrase); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("num_threads", snowflakeCredential.Threads); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSnowflakeCredentialUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	snowflakeCredentialId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("auth_type") || d.HasChange("schema") || d.HasChange("user") || d.HasChange("password") || d.HasChange("num_threads") || d.HasChange("private_key") || d.HasChange("private_key_passphrase") {
		snowflakeCredential, err := c.GetSnowflakeCredential(projectId, snowflakeCredentialId)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("auth_type") {
			authType := d.Get("auth_type").(string)
			snowflakeCredential.Auth_Type = authType
		}
		if d.HasChange("schema") {
			schema := d.Get("schema").(string)
			snowflakeCredential.Schema = schema
		}
		if d.HasChange("user") {
			user := d.Get("user").(string)
			snowflakeCredential.User = user
		}
		if d.HasChange("password") {
			password := d.Get("password").(string)
			snowflakeCredential.Password = password
		}
		if d.HasChange("private_key") {
			privateKey := d.Get("private_key").(string)
			snowflakeCredential.PrivateKey = privateKey
		}
		if d.HasChange("private_key_passphrase") {
			privateKeyPassphrase := d.Get("private_key_passphrase").(string)
			snowflakeCredential.PrivateKeyPassphrase = privateKeyPassphrase
		}
		if d.HasChange("num_threads") {
			numThreads := d.Get("num_threads").(int)
			snowflakeCredential.Threads = numThreads
		}

		_, err = c.UpdateSnowflakeCredential(projectId, snowflakeCredentialId, *snowflakeCredential)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceSnowflakeCredentialRead(ctx, d, m)
}

func resourceSnowflakeCredentialDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	snowflakeCredentialIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	_, err := c.DeleteSnowflakeCredential(snowflakeCredentialIdString, projectIdString)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
