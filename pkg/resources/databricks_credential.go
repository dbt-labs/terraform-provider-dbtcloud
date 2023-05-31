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
	adapterTypes = []string{
		"databricks",
		"spark",
	}
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
				Optional:    true,
				Default:     "default",
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
				Optional:    true,
				Default:     16,
				Description: "Number of threads to use",
			},
			"catalog": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The catalog where to create models (only for the databricks adapter)",
			},
			"schema": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The schema where to create models",
			},
			"adapter_type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of the adapter (databricks or spark)",
				ValidateFunc: validation.StringInSlice(adapterTypes, false),
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
	catalog := d.Get("catalog").(string)
	schema := d.Get("schema").(string)
	adapterType := d.Get("adapter_type").(string)

	databricksCredential, err := c.CreateDatabricksCredential(projectId, "adapter", targetName, adapterId, numThreads, token, catalog, schema, adapterType)
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
	if err := d.Set("catalog", databricksCredential.UnencryptedCredentialDetails["catalog"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schema", databricksCredential.UnencryptedCredentialDetails["schema"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("adapter_type", d.Get("adapter_type").(string)); err != nil {
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

	if d.HasChange("num_threads") || d.HasChange("token") || d.HasChange("target_name") || d.HasChange("catalog") || d.HasChange("schema") || d.HasChange("adapter_type") {
		databricksCredential, err := c.GetDatabricksCredential(projectId, databricksCredentialId)
		if err != nil {
			return diag.FromErr(err)
		}
		if d.HasChange("num_threads") {
			numThreads := d.Get("num_threads").(int)
			databricksCredential.Threads = numThreads
		}
		if d.HasChange("target_name") {
			targetName := d.Get("target_name").(string)
			databricksCredential.Target_Name = targetName
		}

		// we need to fill in the DatabricksCredentialFieldMetadata for all fields, except token if it was not changed
		validation := dbt_cloud.DatabricksCredentialFieldMetadataValidation{
			Required: false,
		}
		tokenMetadata := dbt_cloud.DatabricksCredentialFieldMetadata{
			Label:       "Token",
			Description: "Personalized user token.",
			Field_Type:  "text",
			Encrypt:     true,
			Validation:  validation,
		}
		catalogMetadata := dbt_cloud.DatabricksCredentialFieldMetadata{
			Label:       "Catalog",
			Description: "Catalog name if Unity Catalog is enabled in your Databricks workspace.  Only available in dbt version 1.1 and later.",
			Field_Type:  "text",
			Encrypt:     false,
			Validation:  validation,
		}
		schemaMetadata := dbt_cloud.DatabricksCredentialFieldMetadata{
			Label:       "Schema",
			Description: "User schema.",
			Field_Type:  "text",
			Encrypt:     false,
			Validation:  validation,
		}

		credentialsFieldToken := dbt_cloud.DatabricksCredentialField{
			Metadata: tokenMetadata,
			Value:    d.Get("token").(string),
		}
		credentialsFieldCatalog := dbt_cloud.DatabricksCredentialField{
			Metadata: catalogMetadata,
			Value:    d.Get("catalog").(string),
		}
		credentialsFieldSchema := dbt_cloud.DatabricksCredentialField{
			Metadata: schemaMetadata,
			Value:    d.Get("schema").(string),
		}

		credentialFields := map[string]dbt_cloud.DatabricksCredentialField{}

		// we update token only if it was changed
		if d.HasChange("token") {
			credentialFields["token"] = credentialsFieldToken
		}

		// only databricks accepts a catalog, not spark
		if d.Get("adapter_type").(string) == "databricks" {
			credentialFields["catalog"] = credentialsFieldCatalog
		}

		credentialFields["schema"] = credentialsFieldSchema

		credentialDetails := dbt_cloud.DatabricksCredentialDetails{
			Fields:      credentialFields,
			Field_Order: []string{},
		}

		databricksCredential.Credential_Details = credentialDetails

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

	// those values don't mean anything for delete operation but they are required by the API
	validation := dbt_cloud.DatabricksCredentialFieldMetadataValidation{
		Required: false,
	}
	catalogMetadata := dbt_cloud.DatabricksCredentialFieldMetadata{
		Label:       "Catalog",
		Description: "Catalog name if Unity Catalog is enabled in your Databricks workspace.  Only available in dbt version 1.1 and later.",
		Field_Type:  "text",
		Encrypt:     false,
		Validation:  validation,
	}
	schemaMetadata := dbt_cloud.DatabricksCredentialFieldMetadata{
		Label:       "Schema",
		Description: "User schema.",
		Field_Type:  "text",
		Encrypt:     false,
		Validation:  validation,
	}
	credentialsFieldCatalog := dbt_cloud.DatabricksCredentialField{
		Metadata: catalogMetadata,
		Value:    "NA",
	}
	credentialsFieldSchema := dbt_cloud.DatabricksCredentialField{
		Metadata: schemaMetadata,
		Value:    "NA",
	}
	credentialFields := map[string]dbt_cloud.DatabricksCredentialField{}
	credentialFields["catalog"] = credentialsFieldCatalog
	credentialFields["schema"] = credentialsFieldSchema

	credentialDetails := dbt_cloud.DatabricksCredentialDetails{
		Fields:      credentialFields,
		Field_Order: []string{},
	}

	databricksCredential.Credential_Details = credentialDetails

	_, err = c.UpdateDatabricksCredential(projectId, databricksCredentialId, *databricksCredential)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
