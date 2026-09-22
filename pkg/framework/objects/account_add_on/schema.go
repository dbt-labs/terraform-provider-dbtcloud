package account_add_on

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const spendLimitDescription = "The spend limit set for the add-on, in nanodollars. " +
	"1 USD is 1,000,000,000 nanodollars. Empty when no limit is set. " +
	"The limit works differently for each product. For `wizard` it stops usage: " +
	"the account can no longer use the product once it reaches the limit. For " +
	"`state` nothing is stopped, and the limit only raises a usage alert. " +
	"Change the limit in the dbt platform, because this provider only reads it."

const stateDescription = "The lifecycle state of the add-on: `TRIAL`, `ACTIVE`, " +
	"`CANCELLED` or `EXPIRED`. Empty when the account has never turned the product on. " +
	"A trial moves to `EXPIRED` on its own when it ends"

func (d *accountAddOnDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasource_schema.Schema{
		Description: "Retrieve the state of one usage-based add-on product of the account, " +
			"such as dbt Wizard or dbt State.\n\n" +
			"~> The token needs the billing permission of the account. A product that is " +
			"turned off for the account cannot be read at all, and `wizard` is only " +
			"available while AI features are turned on.",
		Attributes: addOnAttributes(true),
	}
}

func (d *accountAddOnsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasource_schema.Schema{
		Description: "Retrieve the state of every usage-based add-on product the account " +
			"can use, such as dbt Wizard and dbt State.\n\n" +
			"~> The token needs the billing permission of the account. A product that is " +
			"turned off for the account is not in the list.",
		Attributes: map[string]datasource_schema.Attribute{
			"add_ons": datasource_schema.ListNestedAttribute{
				Computed:    true,
				Description: "The add-on products the account can use",
				NestedObject: datasource_schema.NestedAttributeObject{
					Attributes: addOnAttributes(false),
				},
			},
		},
	}
}

// addOnAttributes returns the attributes of one add-on. The single data source
// takes the product from the configuration; the list sets it from the response.
func addOnAttributes(productRequired bool) map[string]datasource_schema.Attribute {
	product := datasource_schema.StringAttribute{
		Computed:    true,
		Description: "The add-on product",
	}
	if productRequired {
		product = datasource_schema.StringAttribute{
			Required:    true,
			Description: "The add-on product to look up. One of `wizard` or `state`",
			Validators: []validator.String{
				stringvalidator.OneOf("wizard", "state"),
			},
		}
	}

	return map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The add-on product",
		},
		"product": product,
		"spend_limit_nanodollars": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: spendLimitDescription,
		},
		"state": datasource_schema.StringAttribute{
			Computed:    true,
			Description: stateDescription,
		},
		"trial_started_at": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "When the trial started",
		},
		"trial_ends_at": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "When the trial ends",
		},
		"trial_consumed": datasource_schema.BoolAttribute{
			Computed:    true,
			Description: "Whether the account has used its one trial of this product",
		},
		"can_activate": datasource_schema.BoolAttribute{
			Computed:    true,
			Description: "Whether the account can start paid use of this product now",
		},
		"can_trial": datasource_schema.BoolAttribute{
			Computed:    true,
			Description: "Whether the account can start a trial of this product now",
		},
	}
}
