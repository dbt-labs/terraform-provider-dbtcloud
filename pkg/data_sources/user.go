package data_sources

import (
	"context"
	"strconv"

	"github.com/gthesheep/terraform-provider-dbt_cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var userSchema = map[string]*schema.Schema{
	"id": &schema.Schema{
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "ID of the user",
	},
	"email": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "Email for the user",
	},
}

func DatasourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceUserRead,
		Schema:      userSchema,
	}
}

func datasourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	email := d.Get("email").(string)

	user, err := c.GetUser(email)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("id", user.ID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(user.ID))

	return diags
}
