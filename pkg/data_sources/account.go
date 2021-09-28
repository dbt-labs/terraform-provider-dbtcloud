package data_sources

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAccountRead,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"plan": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"pending_cancel": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"developer_seats": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"read_only_seats": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"run_slots": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"created_at": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"updated_at": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAccountRead(d *schema.ResourceData, m interface{}) error {
	return nil
}
