package data_sources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var connectionSchema = map[string]*schema.Schema{
	"connection_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID for the connection",
	},
	"project_id": &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "Project ID to create the connection in",
	},
	"is_active": &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the connection is active",
	},
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Connection name",
	},
	"type": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Connection type",
	},
	"private_link_endpoint_id": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ID of the PrivateLink connection",
	},
	"account": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Account for the connection",
	},
	"database": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Database name for the connection",
	},
	"warehouse": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Warehouse name for the connection",
	},
	"role": &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Role name for the connection",
	},
	"allow_sso": &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Flag for whether or not to use SSO for the connection",
	},
	"allow_keep_alive": &schema.Schema{
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Flag for whether or not to use the keep session alive parameter in the connection",
	},
}

func DatasourceConnection() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceConnectionRead,
		Schema:      connectionSchema,
	}
}

func datasourceConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	connectionID := d.Get("connection_id").(int)
	projectID := d.Get("project_id").(int)

	connection, err := c.GetConnection(strconv.Itoa(connectionID), strconv.Itoa(projectID))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("is_active", connection.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", connection.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("connection_id", connection.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", connection.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("account", connection.Details.Account); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", connection.Type); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("private_link_endpoint_id", connection.PrivateLinkEndpointID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("database", connection.Details.Database); err != nil {
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

	d.SetId(fmt.Sprintf("%d%s%d", connection.ProjectID, dbt_cloud.ID_DELIMITER, *connection.ID))

	return diags
}
