package connection_catalog_config

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *connectionCatalogConfigResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(
			`Manages catalog configuration filters for a dbt Cloud connection.
			
This resource configures what database objects (databases, schemas, tables, views) are included 
or excluded when ingesting metadata from your data warehouse. It works in conjunction with 
platform metadata credentials to control what gets synchronized into dbt Cloud's catalog.

Each filter type has an "allow" list (whitelist) and a "deny" list (blacklist):
- If an allow list is set, only matching objects are included
- If a deny list is set, matching objects are excluded
- Deny takes precedence over allow
- Patterns support wildcards (e.g., "temp_*")

~> **Note:** The ~~~connection_id~~~ cannot be changed after creation. To use a different connection, 
you must destroy and recreate the resource.

~> **Note:** This resource requires a platform metadata credential to be configured for the connection.`,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for this resource (connection_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_id": schema.Int64Attribute{
				Description: "The ID of the global connection this catalog config is associated with. Cannot be changed after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"database_allow": schema.ListAttribute{
				Description: "List of database names to include. Supports wildcards (e.g., 'analytics_*'). If set, only these databases are ingested.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"database_deny": schema.ListAttribute{
				Description: "List of database names to exclude. Supports wildcards (e.g., 'staging_*'). Matching databases are not ingested.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"schema_allow": schema.ListAttribute{
				Description: "List of schema names to include. Supports wildcards (e.g., 'public_*'). If set, only these schemas are ingested.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"schema_deny": schema.ListAttribute{
				Description: "List of schema names to exclude. Supports wildcards (e.g., 'temp_*'). Matching schemas are not ingested.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"table_allow": schema.ListAttribute{
				Description: "List of table names to include. Supports wildcards (e.g., 'fact_*'). If set, only these tables are ingested.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"table_deny": schema.ListAttribute{
				Description: "List of table names to exclude. Supports wildcards (e.g., 'tmp_*'). Matching tables are not ingested.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"view_allow": schema.ListAttribute{
				Description: "List of view names to include. Supports wildcards (e.g., 'v_*'). If set, only these views are ingested.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"view_deny": schema.ListAttribute{
				Description: "List of view names to exclude. Supports wildcards (e.g., 'secret_*'). Matching views are not ingested.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}
