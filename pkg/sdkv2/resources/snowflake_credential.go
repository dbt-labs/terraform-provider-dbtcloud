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
			"is_active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the Snowflake credential is active",
			},
			"project_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID to create the Snowflake credential in",
			},
			"credential_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The system Snowflake credential ID",
			},
			"auth_type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of Snowflake credential ('password' or 'keypair')",
				ValidateFunc: validation.StringInSlice(authTypes, false),
			},
			"database": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Database to connect to",
			},
			"role": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Role to assume",
			},
			"warehouse": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Warehouse to use",
			},
			"schema": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Default schema name",
			},
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username for Snowflake",
			},
			"password": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				Description:   "Password for Snowflake",
				ConflictsWith: []string{"private_key", "private_key_passphrase"},
			},
			"private_key": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				Description:   "Private key for Snowflake",
				ConflictsWith: []string{"password"},
			},
			"private_key_passphrase": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				Description:   "Private key passphrase for Snowflake",
				ConflictsWith: []string{"password"},
			},
			"num_threads": {
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

func resourceSnowflakeCredentialCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	authType := d.Get("auth_type").(string)
	database := d.Get("database").(string)
	role := d.Get("role").(string)
	warehouse := d.Get("warehouse").(string)
	schema := d.Get("schema").(string)
	user := d.Get("user").(string)
	password := d.Get("password").(string)
	privateKey := d.Get("private_key").(string)
	privateKeyPassphrase := d.Get("private_key_passphrase").(string)
	numThreads := d.Get("num_threads").(int)

	snowflakeCredential, err := c.CreateSnowflakeCredential(
		projectId,
		"snowflake",
		isActive,
		database,
		role,
		warehouse,
		schema,
		user,
		password,
		privateKey,
		privateKeyPassphrase,
		authType,
		numThreads,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(
		fmt.Sprintf(
			"%d%s%d",
			snowflakeCredential.Project_Id,
			dbt_cloud.ID_DELIMITER,
			*snowflakeCredential.ID,
		),
	)

	resourceSnowflakeCredentialRead(ctx, d, m)

	return diags
}

func resourceSnowflakeCredentialRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	projectId, snowflakeCredentialId, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_snowflake_credential",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	snowflakeCredential, err := c.GetSnowflakeCredential(projectId, snowflakeCredentialId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
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
	if err := d.Set("database", snowflakeCredential.Database); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("role", snowflakeCredential.Role); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("warehouse", snowflakeCredential.Warehouse); err != nil {
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

func resourceSnowflakeCredentialUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, snowflakeCredentialId, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_snowflake_credential",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("auth_type") || d.HasChange("database") || d.HasChange("role") ||
		d.HasChange("warehouse") ||
		d.HasChange("schema") ||
		d.HasChange("user") ||
		d.HasChange("password") ||
		d.HasChange("num_threads") ||
		d.HasChange("private_key") ||
		d.HasChange("private_key_passphrase") {
		snowflakeCredential, err := c.GetSnowflakeCredential(projectId, snowflakeCredentialId)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("auth_type") {
			authType := d.Get("auth_type").(string)
			snowflakeCredential.Auth_Type = authType
		}
		if d.HasChange("database") {
			database := d.Get("database").(string)
			snowflakeCredential.Database = database
		}
		if d.HasChange("role") {
			role := d.Get("role").(string)
			snowflakeCredential.Role = role
		}
		if d.HasChange("warehouse") {
			warehouse := d.Get("warehouse").(string)
			snowflakeCredential.Warehouse = warehouse
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

func resourceSnowflakeCredentialDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString, snowflakeCredentialIdString, err := helper.SplitIDToStrings(
		d.Id(),
		"dbtcloud_snowflake_credential",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.DeleteCredential(snowflakeCredentialIdString, projectIdString)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
