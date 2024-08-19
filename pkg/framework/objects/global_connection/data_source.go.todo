package global_connection

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &globalConnectionDataSource{}
	_ datasource.DataSourceWithConfigure = &globalConnectionDataSource{}
)

func GlobalConnectionDataSource() datasource.DataSource {
	return &globalConnectionDataSource{}
}

type globalConnectionDataSource struct {
	client *dbt_cloud.Client
}

func (d *globalConnectionDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_global_connection"
}

func (d *globalConnectionDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve notification details",
		Attributes:  map[string]schema.Attribute{
			// TODO
		},
	}
}

func (d *globalConnectionDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	// TODO, similar to read resource
}

func (d *globalConnectionDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	switch c := req.ProviderData.(type) {
	case nil: // do nothing
	case *dbt_cloud.Client:
		d.client = c
	default:
		resp.Diagnostics.AddError("Missing client", "A client is required to configure the global connection resource")
	}
}
