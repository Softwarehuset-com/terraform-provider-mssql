package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ServerSchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		Validators: []validator.Object{
			objectvalidator.AtLeastOneOf(path.Expressions{
				path.MatchRelative().AtName("sql_login"),
				path.MatchRelative().AtName("azure_login"),
			}...),
		},
		Required: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required: true,
				Optional: false,
			},
			"port": schema.StringAttribute{
				Required: false,
				Optional: true,
			},
			"sql_login": schema.SingleNestedAttribute{
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.Expressions{
						path.MatchRelative().AtParent().AtName("azure_login"),
					}...),
				},
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Optional: false,
						Required: true,
					},
					"password": schema.StringAttribute{
						Optional: false,
						Required: true,
					},
				},
			},
			"azure_cli": schema.SingleNestedAttribute{
				Optional: true,
			},
			"azure_login": schema.SingleNestedAttribute{

				Optional: true,
				Attributes: map[string]schema.Attribute{
					"tenant_id": schema.StringAttribute{
						Optional: false,
						Required: true,
					},
					"client_id": schema.StringAttribute{
						Optional: false,
						Required: true,
					},
					"client_secret": schema.StringAttribute{
						Optional: false,
						Required: true,
					},
				},
			},
		},
	}
}

func DefaultLanguageSchemaAttribute() schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: "Setting default Language",
	}
}

func DefaultDatabaseSchemaAttribute() schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Description: "Setting default Database",
	}
}
