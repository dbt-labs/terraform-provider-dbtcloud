package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceFabricConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFabricConnectionCreate,
		ReadContext:   resourceFabricConnectionRead,
		UpdateContext: resourceFabricConnectionUpdate,
		DeleteContext: resourceFabricConnectionDelete,

		Description: "Resource to create Microsoft Fabric connections in dbt Cloud",
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Connection Identifier",
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
			// under details
			"server": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The server hostname.",
			},
			"port": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The port to connect to for this connection.",
			},
			"database": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The database to connect to for this connection.",
			},
			"retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The number of automatic times to retry a query before failing. Defaults to 1. Queries with syntax errors will not be retried. This setting can be used to overcome intermittent network issues.",
			},
			"login_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The number of seconds used to establish a connection before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
			},
			"query_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The number of seconds used to wait for a query before failing. Defaults to 0, which means that the timeout is disabled or uses the default system settings.",
			},
			"adapter_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Adapter id created for the Fabric connection",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceFabricConnectionCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m any,
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectId := d.Get("project_id").(int)
	name := d.Get("name").(string)

	server := d.Get("server").(string)
	port := d.Get("port").(int)
	database := d.Get("database").(string)
	retries := d.Get("retries").(int)
	loginTimeout := d.Get("login_timeout").(int)
	queryTimeout := d.Get("query_timeout").(int)

	connection, err := c.CreateFabricConnection(projectId,
		name,
		server,
		port,
		database,
		retries,
		loginTimeout,
		queryTimeout,
	)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", connection.ProjectID, dbt_cloud.ID_DELIMITER, *connection.ID))

	resourceFabricConnectionRead(ctx, d, m)

	return diags
}

func resourceFabricConnectionRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	connectionIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	connection, err := c.GetFabricConnection(connectionIdString, projectIdString)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("connection_id", connection.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", connection.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", connection.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("server", connection.Details.AdapterDetails.Fields["server"].Value.(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("port", int(connection.Details.AdapterDetails.Fields["port"].Value.(float64))); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("database", connection.Details.AdapterDetails.Fields["database"].Value.(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("retries", int(connection.Details.AdapterDetails.Fields["retries"].Value.(float64))); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("login_timeout", int(connection.Details.AdapterDetails.Fields["login_timeout"].Value.(float64))); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("query_timeout", int(connection.Details.AdapterDetails.Fields["query_timeout"].Value.(float64))); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("adapter_id", connection.Details.AdapterId); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceFabricConnectionUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	connectionIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	if d.HasChange("name") ||
		d.HasChange("state") ||
		d.HasChange("server") ||
		d.HasChange("port") ||
		d.HasChange("database") ||
		d.HasChange("retries") ||
		d.HasChange("login_timeout") ||
		d.HasChange("query_timeout") {
		connection, err := c.GetFabricConnection(connectionIdString, projectIdString)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("name") {
			name := d.Get("name").(string)
			connection.Name = name
		}
		if d.HasChange("state") {
			state := d.Get("state").(int)
			connection.State = state
		}

		if d.HasChange("server") || d.HasChange("port") || d.HasChange("database") ||
			d.HasChange("retries") ||
			d.HasChange("login_timeout") ||
			d.HasChange("query_timeout") {
			connection.Details.AdapterDetails = *dbt_cloud.GetFabricConnectionDetails(
				d.Get("server").(string),
				d.Get("port").(int),
				d.Get("database").(string),
				d.Get("retries").(int),
				d.Get("login_timeout").(int),
				d.Get("query_timeout").(int),
			)
		}

		_, err = c.UpdateFabricConnection(connectionIdString, projectIdString, *connection)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceFabricConnectionRead(ctx, d, m)
}

func resourceFabricConnectionDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {

	return resourceConnectionDelete(ctx, d, m)

}
