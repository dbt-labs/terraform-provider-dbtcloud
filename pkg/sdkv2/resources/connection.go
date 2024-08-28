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
	connectionTypes = []string{
		"snowflake",
		"bigquery",
		"redshift",
		"postgres",
		"alloydb",
		"adapter",
	}
)

func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectionCreate,
		ReadContext:   resourceConnectionRead,
		UpdateContext: resourceConnectionUpdate,
		DeleteContext: resourceConnectionDelete,

		Description: helper.DocString(
			`Resource to create a Data Warehouse connection in dbt Cloud.
			
			~> This resource is deprecated and is going to be removed in the next major release, please use the ~~~dbtcloud_global_connection~~~ resource instead to create connections.`,
		),
		DeprecationMessage: "Please replace this resource with a `dbtcloud_global_connection` resource. This resource type will be removed in the next major release.",

		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Connection Identifier",
			},
			"is_active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the connection is active",
			},
			"project_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID to create the connection in",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Connection name",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of connection",
				ValidateFunc: validation.StringInSlice(connectionTypes, false),
				ForceNew:     true,
			},
			"private_link_endpoint_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     nil,
				Description: "The ID of the PrivateLink connection. This ID can be found using the `privatelink_endpoint` data source",
			},
			"account": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Account name for the connection (for Snowflake)",
				ConflictsWith: []string{"host_name"},
			},
			"host_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Host name for the connection, including Databricks cluster",
				ConflictsWith: []string{"account"},
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     nil,
				Description: "Port number to connect via",
			},
			"database": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Database name for the connection",
			},
			"warehouse": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Warehouse name for the connection (for Snowflake)",
			},
			"role": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Role name for the connection (for Snowflake)",
			},
			"allow_sso": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not the connection should allow SSO (for Snowflake)",
			},
			"allow_keep_alive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not the connection should allow client session keep alive (for Snowflake)",
			},
			"oauth_client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "OAuth client identifier (for Snowflake and Databricks)",
			},
			"oauth_client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "OAuth client secret (for Snowflake and Databricks)",
			},
			"tunnel_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not tunneling should be enabled on your database connection",
			},
			"http_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The HTTP path of the Databricks cluster or SQL warehouse (for Databricks)",
			},
			"catalog": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Catalog name if Unity Catalog is enabled in your Databricks workspace (for Databricks)",
			},
			"adapter_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Adapter id created for the Databricks connection (for Databricks)",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceConnectionCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	name := d.Get("name").(string)
	connectionType := d.Get("type").(string)
	privatelinkEndpointID := d.Get("private_link_endpoint_id").(string)
	account := d.Get("account").(string)
	database := d.Get("database").(string)
	warehouse := d.Get("warehouse").(string)
	role := d.Get("role").(string)
	allowSSO := d.Get("allow_sso").(bool)
	allowKeepAlive := d.Get("allow_keep_alive").(bool)
	oAuthClientID := d.Get("oauth_client_id").(string)
	oAuthClientSecret := d.Get("oauth_client_secret").(string)
	hostName := d.Get("host_name").(string)
	port := d.Get("port").(int)
	tunnelEnabled := d.Get("tunnel_enabled").(bool)
	httpPath := d.Get("http_path").(string)
	catalog := d.Get("catalog").(string)

	connection, err := c.CreateConnection(
		projectId,
		name,
		connectionType,
		privatelinkEndpointID,
		isActive,
		account,
		database,
		warehouse,
		role,
		&allowSSO,
		&allowKeepAlive,
		oAuthClientID,
		oAuthClientSecret,
		hostName,
		port,
		&tunnelEnabled,
		httpPath,
		catalog,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", connection.ProjectID, dbt_cloud.ID_DELIMITER, *connection.ID))

	resourceConnectionRead(ctx, d, m)

	return diags
}

func resourceConnectionRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	connectionIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	connection, err := c.GetConnection(connectionIdString, projectIdString)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// TODO: Remove when done better
	connection.Details.OAuthClientID = d.Get("oauth_client_id").(string)
	connection.Details.OAuthClientSecret = d.Get("oauth_client_secret").(string)

	if err := d.Set("connection_id", connection.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_active", connection.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", connection.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", connection.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", connection.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("private_link_endpoint_id", connection.PrivateLinkEndpointID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("account", connection.Details.Account); err != nil {
		return diag.FromErr(err)
	}

	databaseName := connection.Details.DBName
	if connection.Type == "snowflake" {
		databaseName = connection.Details.Database
	}
	if err := d.Set("database", databaseName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("warehouse", connection.Details.Warehouse); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("role", connection.Details.Role); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("allow_sso", connection.Details.AllowSSO); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("allow_keep_alive", connection.Details.ClientSessionKeepAlive); err != nil {
		return diag.FromErr(err)
	}
	if d.Get("type") == "snowflake" {
		if err := d.Set("oauth_client_id", connection.Details.OAuthClientID); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("oauth_client_secret", connection.Details.OAuthClientSecret); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("port", connection.Details.Port); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tunnel_enabled", connection.Details.TunnelEnabled); err != nil {
		return diag.FromErr(err)
	}
	httpPath := ""
	catalog := ""
	hostName := connection.Details.Host
	clientID := ""
	clientSecret := ""
	if connection.Details.AdapterDetails != nil {
		httpPath = connection.Details.AdapterDetails.Fields["http_path"].Value.(string)
		catalog = connection.Details.AdapterDetails.Fields["catalog"].Value.(string)
		hostName = connection.Details.AdapterDetails.Fields["host"].Value.(string)
		clientID = d.Get("oauth_client_id").(string)
		clientSecret = d.Get("oauth_client_secret").(string)
	}
	if err := d.Set("host_name", hostName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("http_path", httpPath); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("catalog", catalog); err != nil {
		return diag.FromErr(err)
	}
	// we set those just for the adapter as the logic for Snowflake is up in this function
	if d.Get("type") == "adapter" {
		if err := d.Set("oauth_client_id", clientID); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("oauth_client_secret", clientSecret); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("adapter_id", connection.Details.AdapterId); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceConnectionUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	connectionIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	if d.HasChange("name") ||
		d.HasChange("type") ||
		d.HasChange("private_link_endpoint_id") ||
		d.HasChange("account") ||
		d.HasChange("host_name") ||
		d.HasChange("port") ||
		d.HasChange("database") ||
		d.HasChange("warehouse") ||
		d.HasChange("role") ||
		d.HasChange("allow_sso") ||
		d.HasChange("allow_keep_alive") ||
		d.HasChange("oauth_client_id") ||
		d.HasChange("oauth_client_secret") ||
		d.HasChange("tunnel_enabled") ||
		d.HasChange("http_path") ||
		d.HasChange("catalog") ||
		d.HasChange("adapter_id") {
		connection, err := c.GetConnection(connectionIdString, projectIdString)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("name") {
			name := d.Get("name").(string)
			connection.Name = name
		}
		if d.HasChange("type") {
			connectionType := d.Get("type").(string)
			connection.Type = connectionType
		}
		if d.HasChange("private_link_endpoint_id") {
			privatelinkEndpointID := d.Get("private_link_endpoint_id").(string)
			connection.PrivateLinkEndpointID = privatelinkEndpointID
		}
		if d.HasChange("account") {
			account := d.Get("account").(string)
			connection.Details.Account = account
		}
		if d.HasChange("host_name") {
			hostName := d.Get("host_name").(string)
			connection.Details.Host = hostName
		}
		if d.HasChange("port") {
			port := d.Get("port").(int)
			connection.Details.Port = port
		}
		if d.HasChange("database") {
			database := d.Get("database").(string)
			if connection.Type == "snowflake" {
				connection.Details.Database = database
			} else {
				connection.Details.DBName = database
			}
		}
		if d.HasChange("warehouse") {
			warehouse := d.Get("warehouse").(string)
			connection.Details.Warehouse = warehouse
		}
		if d.HasChange("role") {
			role := d.Get("role").(string)
			connection.Details.Role = role
		}
		if d.HasChange("allow_sso") {
			allowSSO := d.Get("allow_sso").(bool)
			connection.Details.AllowSSO = &allowSSO
		}
		if d.HasChange("allow_keep_alive") {
			allowKeepAlive := d.Get("allow_keep_alive").(bool)
			connection.Details.ClientSessionKeepAlive = &allowKeepAlive
		}
		if d.HasChange("oauth_client_id") && d.Get("type") == "snowflake" {
			oAuthClientID := d.Get("oauth_client_id").(string)
			connection.Details.OAuthClientID = oAuthClientID
		}
		if d.HasChange("oauth_client_secret") && d.Get("type") == "snowflake" {
			oAuthClientSecret := d.Get("oauth_client_secret").(string)
			connection.Details.OAuthClientSecret = oAuthClientSecret
		}
		if d.HasChange("tunnel_enabled") {
			tunnelEnabled := d.Get("tunnel_enabled").(bool)
			connection.Details.TunnelEnabled = &tunnelEnabled
		}
		if d.Get("type") == "adapter" &&
			(d.HasChange("http_path") || d.HasChange("host_name") || d.HasChange("catalog") ||
				d.HasChange("oauth_client_id") ||
				d.HasChange("oauth_client_secret")) {
			connection.Details.AdapterDetails = dbt_cloud.GetDatabricksConnectionDetails(
				d.Get("host_name").(string),
				d.Get("http_path").(string),
				d.Get("catalog").(string),
				d.Get("oauth_client_id").(string),
				d.Get("oauth_client_secret").(string),
			)
		}
		if d.HasChange("adapter_id") {
			adapterId := d.Get("adapter_id").(int)
			connection.Details.AdapterId = &adapterId
		}

		_, err = c.UpdateConnection(connectionIdString, projectIdString, *connection)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceConnectionRead(ctx, d, m)
}

func resourceConnectionDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	connectionIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	_, err := c.DeleteConnection(connectionIdString, projectIdString)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
