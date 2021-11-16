package data_sources_test

import (
	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func providers() map[string]*schema.Provider {
	p := provider.Provider()
	return map[string]*schema.Provider{
		"dbt": p,
	}
}
