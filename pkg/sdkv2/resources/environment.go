package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEnvironmentCreate,
		ReadContext:   resourceEnvironmentRead,
		UpdateContext: resourceEnvironmentUpdate,
		DeleteContext: resourceEnvironmentDelete,
		Description: helper.DocString(
			`Resource to manage dbt Cloud environments for the different dbt Cloud projects.

			In a given dbt Cloud project, one development environment can be defined and as many deployment environments as needed can be created.

			~> In August 2024, dbt Cloud released the "global connection" feature, allowing connections to be defined at the account level and reused across environments and projects.
			This version of the provider has the ~~~connection_id~~~ as an optional field but it is recommended to start setting it up in your projects. In future versions, this field will become mandatory.
			`,
		),

		Schema: map[string]*schema.Schema{
			"is_active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the environment is active",
			},
			"project_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Project ID to create the environment in",
			},
			"credential_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     nil,
				Description: "Credential ID to create the environment with. A credential is not required for development environments but is required for deployment environments",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Environment name",
			},
			"dbt_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "latest",
				Description: "Version number of dbt to use in this environment. It needs to be in the format `major.minor.0-latest` (e.g. `1.5.0-latest`), `major.minor.0-pre`, `versionless`, or `latest`. While `versionless` is still supported, using `latest` is recommended. Defaults to `latest` if no version is provided",
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					switch oldValue {
					case "versionless":
						return newValue == "latest"
					case "latest":
						return newValue == "versionless"
					default:
						return false
					}
				},
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of environment (must be either development or deployment)",
				ForceNew:    true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					type_ := val.(string)
					switch type_ {
					case
						"development",
						"deployment":
						return
					}
					errs = append(
						errs,
						fmt.Errorf(
							"%q must be either development or deployment, got: %q",
							key,
							type_,
						),
					)
					return
				},
			},
			"use_custom_branch": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to use a custom git branch in this environment",
			},
			"custom_branch": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Which custom branch to use in this environment",
			},
			"deployment_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The type of environment. Only valid for environments of type 'deployment' and for now can only be 'production', 'staging' or left empty for generic environments",
			},
			"environment_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Environment ID within the project",
			},
			"extended_attributes_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ID of the extended attributes for the environment",
			},
			"connection_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Description: helper.DocString(
					`The ID of the connection to use (can be the ~~~id~~~ of a ~~~dbtcloud_global_connection~~~ or the ~~~connection_id~~~ of a legacy connection). 
					  - At the moment, it is optional and the environment will use the connection set in ~~~dbtcloud_project_connection~~~ if ~~~connection_id~~~ is not set in this resource
					  - In future versions this field will become required, so it is recommended to set it from now on
					  - When configuring this field, it needs to be configured for all the environments of the project
					  - To avoid Terraform state issues, when using this field, the ~~~dbtcloud_project_connection~~~ resource should be removed from the project or you need to make sure that the ~~~connection_id~~~ is the same in ~~~dbtcloud_project_connection~~~ and in the ~~~connection_id~~~ of the Development environment of the project`,
				),
			},
			"enable_model_query_history": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to enable model query history in this environment. As of Oct 2024, works only for Snowflake and BigQuery.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceEnvironmentCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	credentialId := d.Get("credential_id").(int)
	name := d.Get("name").(string)
	dbtVersion := d.Get("dbt_version").(string)
	type_ := d.Get("type").(string)
	useCustomBranch := d.Get("use_custom_branch").(bool)
	customBranch := d.Get("custom_branch").(string)
	deploymentType := d.Get("deployment_type").(string)
	extendedAttributesID := d.Get("extended_attributes_id").(int)
	connectionID := d.Get("connection_id").(int)
	enableModelQueryHistory := d.Get("enable_model_query_history").(bool)

	environment, err := c.CreateEnvironment(
		isActive,
		projectId,
		name,
		dbtVersion,
		type_,
		useCustomBranch,
		customBranch,
		credentialId,
		deploymentType,
		extendedAttributesID,
		connectionID,
		enableModelQueryHistory,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", environment.Project_Id, dbt_cloud.ID_DELIMITER, *environment.ID))

	resourceEnvironmentRead(ctx, d, m)

	return diags
}

func resourceEnvironmentRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectId, environmentId, err := helper.SplitIDToInts(d.Id(), "dbtcloud_environment")
	if err != nil {
		return diag.FromErr(err)
	}

	environment, err := c.GetEnvironment(projectId, environmentId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("is_active", environment.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", environment.Project_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", environment.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("dbt_version", environment.Dbt_Version); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", environment.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("use_custom_branch", environment.Use_Custom_Branch); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("custom_branch", environment.Custom_Branch); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("environment_id", environment.Environment_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("credential_id", environment.Credential_Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("deployment_type", environment.DeploymentType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("extended_attributes_id", environment.ExtendedAttributesID); err != nil {
		return diag.FromErr(err)
	}
	if v, ok := d.GetOk("connection_id"); ok && v != nil {
		if err := d.Set("connection_id", environment.ConnectionID); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("connection_id", 0); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("enable_model_query_history", environment.EnableModelQueryHistory); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceEnvironmentUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectId, environmentId, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_environment",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("name") ||
		d.HasChange("dbt_version") ||
		d.HasChange("credential_id") ||
		d.HasChange("project_id") ||
		d.HasChange("type") ||
		d.HasChange("custom_branch") ||
		d.HasChange("use_custom_branch") ||
		d.HasChange("deployment_type") ||
		d.HasChange("extended_attributes_id") ||
		d.HasChange("connection_id") ||
		d.HasChange("enable_model_query_history") {

		environment, err := c.GetEnvironment(projectId, environmentId)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("name") {
			name := d.Get("name").(string)
			environment.Name = name
		}
		if d.HasChange("dbt_version") {
			Dbt_Version := d.Get("dbt_version").(string)
			environment.Dbt_Version = Dbt_Version
		}
		if d.HasChange("credential_id") {
			Credential_Id := d.Get("credential_id").(int)
			environment.Credential_Id = &Credential_Id
		}
		if d.HasChange("project_id") {
			Project_Id := d.Get("project_id").(int)
			environment.Project_Id = Project_Id
		}
		if d.HasChange("type") {
			Type := d.Get("type").(string)
			environment.Type = Type
		}
		if d.HasChange("custom_branch") {
			Custom_Branch := d.Get("custom_branch").(string)
			environment.Custom_Branch = &Custom_Branch
		}
		if d.HasChange("use_custom_branch") {
			Use_Custom_Branch := d.Get("use_custom_branch").(bool)
			environment.Use_Custom_Branch = Use_Custom_Branch
		}
		if d.HasChange("deployment_type") {
			DeploymentType := d.Get("deployment_type").(string)
			if DeploymentType != "" {
				environment.DeploymentType = &DeploymentType
			} else {
				environment.DeploymentType = nil
			}
		}
		if d.HasChange("extended_attributes_id") {
			extendedAttributesID := d.Get("extended_attributes_id").(int)
			if extendedAttributesID != 0 {
				environment.ExtendedAttributesID = &extendedAttributesID
			} else {
				environment.ExtendedAttributesID = nil
			}
		}
		if d.HasChange("connection_id") {
			connectionID := d.Get("connection_id").(int)
			if connectionID != 0 {
				environment.ConnectionID = &connectionID
			} else {
				environment.ConnectionID = nil
			}
		}
		if d.HasChange("enable_model_query_history") {
			environment.EnableModelQueryHistory = d.Get("enable_model_query_history").(bool)
		}
		_, err = c.UpdateEnvironment(projectId, environmentId, *environment)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceEnvironmentRead(ctx, d, m)
}

func resourceEnvironmentDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectId, environmentId, err := helper.SplitIDToInts(
		d.Id(),
		"dbtcloud_environment",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.DeleteEnvironment(projectId, environmentId)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
