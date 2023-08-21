// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	planmodifier "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/softwarehuset/mssql/internal/model"
	"github.com/softwarehuset/mssql/internal/sql"
	"net/http"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

// AadLoginResource defines the resource implementation.
type UserResource struct {
	client *http.Client
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_user_v2"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Example resource",

		Attributes: map[string]schema.Attribute{

			"username": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "Example configurable attribute",
				Optional:            false,
				Required:            true,
			},
			"login_name": schema.StringAttribute{
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "Example configurable attribute",
				Optional:            false,
				Required:            true,
			},
			"roles": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Example configurable attribute",
				Optional:            false,
				Required:            true,
			},
			"database": schema.StringAttribute{
				MarkdownDescription: "Example configurable attribute",
				Optional:            false,
				Required:            true,
			},
			"default_schema": schema.StringAttribute{
				MarkdownDescription: "Example configurable attribute",
				Optional:            true,
				Required:            false,
				Computed:            true,
				Default:             stringdefault.StaticString("dbo"),
			},
			"default_language": DefaultLanguageSchemaAttribute(),
			"server":           ServerSchemaAttribute(),
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data model.UserResourceModel
	tflog.Info(ctx, fmt.Sprintf("CreateRequest: %v", data))
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var authType string
	if data.LoginName.ValueString() != "" {
		authType = "INSTANCE"
	} else {
		authType = "EXTERNAL"
	}

	var sqlClient, _ = sql.GetFactory().GetSqlClient(ctx, *data.Server, data.Database.ValueString())

	var sqlData = model.User{
		UserName:        data.UserName,
		LoginName:       data.LoginName,
		DefaultSchema:   data.DefaultSchema,
		DefaultLanguage: data.DefaultLanguage,
		Database:        data.Database,
		Roles:           data.Roles,
		AuthType:        basetypes.NewStringValue(authType),
	}
	err := sqlClient.(sql.UserConnector).CreateUser(ctx, data.Database.ValueString(), &sqlData)

	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError(
			"API Error Creating Resource",
			fmt.Sprintf("... details ... %s", err))

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data model.UserResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data model.User

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var sqlClient, _ = GetSqlClientFactory().GetSqlClient(ctx, *data.Server, data.Database.ValueString())
	err := sqlClient.(sql.UserConnector).UpdateUser(ctx, data.Database.ValueString(), &data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))

		return
	}

	if sqlClient == nil {

	}
	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data model.UserResourceModel
	tflog.Info(ctx, fmt.Sprintf("CreateRequest: %v", data))
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var sqlClient, _ = sql.GetFactory().GetSqlClient(ctx, *data.Server, data.Database.ValueString())
	err := sqlClient.(sql.UserConnector).DeleteUser(ctx, data.Database.ValueString(), data.UserName.ValueString())
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError(
			"API Error Deleting Resource",
			fmt.Sprintf("... details ... %s", err))

		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}
