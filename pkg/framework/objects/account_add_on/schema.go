package account_add_on

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const spendLimitDescription = "The spend limit for the add-on, in nanodollars. " +
	"1 USD is 1,000,000,000 nanodollars. Leave it out for no limit. " +
	"The limit works differently for each product. For `wizard` it stops usage: " +
	"the account can no longer use the product once it reaches the limit, and the " +
	"limit can not be set below the amount already spent in the current billing " +
	"period. For `state` nothing is stopped, and the limit only raises a usage alert."

func (r *accountAddOnResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resource_schema.Schema{
		Description: "Manages a usage-based add-on product on the account, such as dbt Wizard " +
			"or dbt State.\n\n" +
			"~> **Billing** Setting `activation` to `paid` starts paid use of the product, " +
			"and the account is charged for the usage. If the account is not on a " +
			"usage-based plan yet, the activation also moves the account to one. Read the " +
			"plan change in the apply output before you accept it. Destroying this resource " +
			"cancels the product immediately.\n\n" +
			"~> **Permissions** The token needs the billing permission of the account. " +
			"`wizard` is only available while AI features are turned on for the account, " +
			"which you can manage with `dbtcloud_account_features`.",
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.StringAttribute{
				Computed:    true,
				Description: "The add-on product, which is also the import ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"product": resource_schema.StringAttribute{
				Required:    true,
				Description: "The add-on product to manage. One of `wizard` or `state`",
				Validators: []validator.String{
					stringvalidator.OneOf("wizard", "state"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"activation": resource_schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("paid"),
				Description: "How to turn the product on. `paid` starts paid use and charges " +
					"the account. `trial` starts the trial instead, and an account gets one " +
					"trial for each product. The value is only used when the resource is " +
					"created, and a change to it replaces the resource",
				Validators: []validator.String{
					stringvalidator.OneOf("paid", "trial"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"spend_limit_nanodollars": resource_schema.Int64Attribute{
				Optional:    true,
				Description: spendLimitDescription,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"state": resource_schema.StringAttribute{
				Computed: true,
				Description: "The lifecycle state of the add-on: `TRIAL`, `ACTIVE`, " +
					"`CANCELLED` or `EXPIRED`. A trial moves to `EXPIRED` on its own when it " +
					"ends, so the value can change without a change to the configuration. " +
					"The plan stays empty in that case, and the provider raises a warning " +
					"instead, because the state is read-only. To use a product again after " +
					"it ends, replace the resource",
			},
			"trial_started_at": resource_schema.StringAttribute{
				Computed:    true,
				Description: "When the trial started",
			},
			"trial_ends_at": resource_schema.StringAttribute{
				Computed:    true,
				Description: "When the trial ends",
			},
			"trial_consumed": resource_schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the account has used its one trial of this product",
			},
			"can_activate": resource_schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the account can start paid use of this product now",
			},
			"can_trial": resource_schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the account can start a trial of this product now",
			},
		},
	}
}

func (d *accountAddOnDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasource_schema.Schema{
		Description: "Retrieve the state of one usage-based add-on product on the account.",
		Attributes:  addOnDataSourceAttributes(true),
	}
}

func (d *accountAddOnsDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasource_schema.Schema{
		Description: "Retrieve the state of every usage-based add-on product the account can use. " +
			"A product that is turned off for the account is not in the list.",
		Attributes: map[string]datasource_schema.Attribute{
			"add_ons": datasource_schema.ListNestedAttribute{
				Computed:    true,
				Description: "The add-on products the account can use",
				NestedObject: datasource_schema.NestedAttributeObject{
					Attributes: addOnDataSourceAttributes(false),
				},
			},
		},
	}
}

// addOnDataSourceAttributes returns the attributes of one add-on. The single data
// source takes the product from the configuration; the list sets it from the response.
func addOnDataSourceAttributes(productRequired bool) map[string]datasource_schema.Attribute {
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
		"id":      idAttribute(),
		"product": product,
		"spend_limit_nanodollars": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: spendLimitDescription,
		},
		"state": datasource_schema.StringAttribute{
			Computed: true,
			Description: "The lifecycle state of the add-on: `TRIAL`, `ACTIVE`, `CANCELLED` " +
				"or `EXPIRED`. Empty when the account has never used the product",
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

func idAttribute() datasource_schema.Attribute {
	return datasource_schema.StringAttribute{
		Computed:    true,
		Description: "The add-on product",
	}
}
