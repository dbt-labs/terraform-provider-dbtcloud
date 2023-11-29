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

func ResourceFabricCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFabricCredentialCreate,
		ReadContext:   resourceFabricCredentialRead,
		UpdateContext: resourceFabricCredentialUpdate,
		DeleteContext: resourceFabricCredentialDelete,

		Schema: map[string]*schema.Schema{
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID to create the Fabric credential in",
			},
			"adapter_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Fabric adapter ID for the credential",
			},
			"credential_id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The system Fabric credential ID",
			},
			"user": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Description:   "The username of the Fabric account to connect to. Only used when connection with AD user/pass",
				ConflictsWith: []string{"tenant_id", "client_id", "client_secret"},
			},
			"password": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				Default:       "",
				Description:   "The password for the account to connect to. Only used when connection with AD user/pass",
				ConflictsWith: []string{"tenant_id", "client_id", "client_secret"},
			},
			"tenant_id": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Description:   "The tenant ID of the Azure Active Directory instance. This is only used when connecting to Azure SQL with a service principal.",
				ConflictsWith: []string{"user", "password"},
			},
			"client_id": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Description:   "The client ID of the Azure Active Directory service principal. This is only used when connecting to Azure SQL with an AAD service principal.",
				ConflictsWith: []string{"user", "password"},
			},
			"client_secret": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				Default:       "",
				Description:   "The client secret of the Azure Active Directory service principal. This is only used when connecting to Azure SQL with an AAD service principal.",
				ConflictsWith: []string{"user", "password"},
			},
			"schema": &schema.Schema{
				Type:          schema.TypeString,
				Required:      true,
				Description:   "The schema where to create the dbt models",
				ConflictsWith: []string{},
			},
			"schema_authorization": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Description:   "Optionally set this to the principal who should own the schemas created by dbt",
				ConflictsWith: []string{},
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceFabricCredentialCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId := d.Get("project_id").(int)
	adapterId := d.Get("adapter_id").(int)
	user := d.Get("user").(string)
	password := d.Get("password").(string)
	tenantId := d.Get("tenant_id").(string)
	clientId := d.Get("client_id").(string)
	clientSecret := d.Get("client_secret").(string)
	schema := d.Get("schema").(string)
	schemaAuthorization := d.Get("schema_authorization").(string)

	// FRAMEWORK: Move this logic to the schema validation when moving to the Framework
	userPasswordDefined := user != "" && password != ""
	servicePrincipalDefined := tenantId != "" && clientId != "" && clientSecret != ""

	if !userPasswordDefined && !servicePrincipalDefined {
		diag.FromErr(fmt.Errorf("either user/password or service principal auth must be defined"))
	}

	fabricCredential, err := c.CreateFabricCredential(
		projectId,
		adapterId,
		user,
		password,
		tenantId,
		clientId,
		clientSecret,
		schema,
		schemaAuthorization,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(
		fmt.Sprintf(
			"%d%s%d",
			fabricCredential.Project_Id,
			dbt_cloud.ID_DELIMITER,
			*fabricCredential.ID,
		),
	)

	resourceFabricCredentialRead(ctx, d, m)

	return diags
}

func resourceFabricCredentialRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	fabricCredentialId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	fabricCredential, err := c.GetFabricCredential(projectId, fabricCredentialId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// set the ones that don't come back from the API

	if err := d.Set("project_id", fabricCredential.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("adapter_id", fabricCredential.Adapter_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("credential_id", fabricCredentialId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user", fabricCredential.UnencryptedCredentialDetails.User); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tenant_id", fabricCredential.UnencryptedCredentialDetails.TenantId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("client_id", fabricCredential.UnencryptedCredentialDetails.ClientId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schema", fabricCredential.UnencryptedCredentialDetails.Schema); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schema_authorization", fabricCredential.UnencryptedCredentialDetails.SchemaAuthorization); err != nil {
		return diag.FromErr(err)
	}

	// set the ones that don't come back from the API
	if err := d.Set("password", d.Get("password").(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("client_secret", d.Get("client_secret").(string)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceFabricCredentialUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}

	fabricCredentialId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("adapter_id") || d.HasChange("user") ||
		d.HasChange("password") ||
		d.HasChange("tenant_id") ||
		d.HasChange("client_id") ||
		d.HasChange("client_secret") ||
		d.HasChange("schema") ||
		d.HasChange("schema_authorization") {

		adapterId := d.Get("adapter_id").(int)
		user := d.Get("user").(string)
		password := d.Get("password").(string)
		tenantId := d.Get("tenant_id").(string)
		clientId := d.Get("client_id").(string)
		clientSecret := d.Get("client_secret").(string)
		schema := d.Get("schema").(string)
		schemaAuthorization := d.Get("schema_authorization").(string)

		// FRAMEWORK: Move this logic to the schema validation when moving to the Framework
		userPasswordDefined := user != "" && password != ""
		servicePrincipalDefined := tenantId != "" && clientId != "" && clientSecret != ""

		if !userPasswordDefined && !servicePrincipalDefined {
			return diag.FromErr(
				fmt.Errorf("either user/password or service principal auth must be defined"),
			)
		}

		fabricCredential, err := c.GetFabricCredential(projectId, fabricCredentialId)
		if err != nil {
			return diag.FromErr(err)
		}
		if d.HasChange("adapter_id") {
			fabricCredential.Adapter_Id = adapterId
		}

		fabricCredentialDetails, err := dbt_cloud.GenerateFabricCredentialDetails(
			user,
			password,
			tenantId,
			clientId,
			clientSecret,
			schema,
			schemaAuthorization,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		fabricCredential.CredentialDetails = fabricCredentialDetails

		_, err = c.UpdateFabricCredential(
			projectId,
			fabricCredentialId,
			*fabricCredential,
		)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceFabricCredentialRead(ctx, d, m)
}

func resourceFabricCredentialDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0])
	if err != nil {
		return diag.FromErr(err)
	}
	fabricCredentialId, err := strconv.Atoi(strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1])
	if err != nil {
		return diag.FromErr(err)
	}

	fabricCredential, err := c.GetFabricCredential(projectId, fabricCredentialId)
	if err != nil {
		return diag.FromErr(err)
	}

	fabricCredential.State = dbt_cloud.STATE_DELETED

	// // those values don't mean anything for the delete operation but they are required by the API
	emptyFabricCredentialDetails, err := dbt_cloud.GenerateFabricCredentialDetails(
		"",
		"",
		"",
		"",
		"",
		"",
		"",
	)

	fabricCredential.CredentialDetails = emptyFabricCredentialDetails

	_, err = c.UpdateFabricCredential(projectId, fabricCredentialId, *fabricCredential)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
