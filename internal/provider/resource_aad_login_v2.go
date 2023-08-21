// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/softwarehuset/mssql/internal/model"
	"github.com/softwarehuset/mssql/internal/sql"
	"net/http"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AadLoginResource{}

func NewAadLoginResource() resource.Resource {
	return &AadLoginResource{}
}

// AadLoginResource defines the resource implementation.
type AadLoginResource struct {
	client *http.Client
}

func (r *AadLoginResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aad_login_v2"
}

func (r *AadLoginResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example resource",

		Attributes: map[string]schema.Attribute{
			"aad_login_name": schema.StringAttribute{
				MarkdownDescription: "Example configurable attribute",
				Optional:            false,
				Required:            true,
			},
			"default_database": DefaultDatabaseSchemaAttribute(),
			"default_language": DefaultLanguageSchemaAttribute(),
			"server":           ServerSchemaAttribute(),
		},
	}
}

func (r *AadLoginResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AadLoginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data model.AadLogin
	tflog.Info(ctx, fmt.Sprintf("CreateRequest: %v", data))
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var sqlClient, _ = sql.GetFactory().GetSqlClient(ctx, *data.Server, "master")
	err := sqlClient.(sql.AadLoginConnector).CreateAadLogin(ctx, data.AadLoginName.ValueString(), data.DefaultDatabase.ValueString(), data.DefaultLanguage.ValueString())
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError(
			"API Error Creating Resource",
			fmt.Sprintf("... details ... %s", err))

		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AadLoginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data model.AadLogin

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

func (r *AadLoginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data model.AadLogin

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var sqlClient, _ = GetSqlClientFactory().GetSqlClient(ctx, *data.Server, "master")
	sqlClient.(sql.AadLoginConnector).CreateAadLogin(ctx, data.AadLoginName.String(), data.DefaultDatabase.String(), data.DefaultLanguage.String())

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

func (r *AadLoginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data model.AadLogin
	tflog.Info(ctx, fmt.Sprintf("CreateRequest: %v", data))
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	var sqlClient, _ = sql.GetFactory().GetSqlClient(ctx, *data.Server, "master")
	err := sqlClient.(sql.AadLoginConnector).DeleteAadLogin(ctx, data.AadLoginName.ValueString())
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
