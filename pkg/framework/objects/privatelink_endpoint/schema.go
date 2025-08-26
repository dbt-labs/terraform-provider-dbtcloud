package privatelink_endpoint

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var datasourceSchema = datasource_schema.Schema{
	Description: "Privatelink endpoint data source.",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The internal ID of the PrivateLink Endpoint",
		},
		"name": datasource_schema.StringAttribute{
			Optional:    true,
			Description: "Given descriptive name for the PrivateLink Endpoint (name and/or private_link_endpoint_url need to be provided to return data for the datasource)",
		},
		"type": datasource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Type of the PrivateLink Endpoint",
		},
		"private_link_endpoint_url": datasource_schema.StringAttribute{
			Optional:    true,
			Description: "URL of the PrivateLink Endpoint (name and/or private_link_endpoint_url need to be provided to return data for the datasource)",
		},
		"cidr_range": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "CIDR range of the PrivateLink Endpoint",
		},
	},
}

func (r *privatelinkEndpointDataSourceAll) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	var datasourceSchemaAll = datasource_schema.Schema{
		Description: "Privatelink endpoint data sources.",
		Attributes: map[string]datasource_schema.Attribute{
			"id": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "The internal ID of the PrivateLink Endpoint",
			},
		},
	}

	resp.Schema = datasourceSchemaAll
}