package model

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SqlClientFactory interface {
	GetSqlClient(ctx context.Context, server Server, database string) (interface{}, error)
}

type UserResourceModel struct {
	UserName        types.String `tfsdk:"username"`
	LoginName       types.String `tfsdk:"login_name"`
	DefaultSchema   types.String `tfsdk:"default_schema"`
	DefaultLanguage types.String `tfsdk:"default_language"`
	Database        types.String `tfsdk:"database"`
	Roles           types.List   `tfsdk:"roles"`
	Server          *Server      `tfsdk:"server"`
}

type User struct {
	PrincipalID     types.Int64  `tfsdk:"principal_id"`
	UserName        types.String `tfsdk:"username"`
	ObjectId        types.String `tfsdk:"object_id"`
	LoginName       types.String `tfsdk:"login_name"`
	Password        types.String `tfsdk:"password"`
	SIDStr          types.String `tfsdk:"sid_str"`
	AuthType        types.String `tfsdk:"auth_type"`
	DefaultSchema   types.String `tfsdk:"default_schema"`
	DefaultLanguage types.String `tfsdk:"default_language"`
	Database        types.String `tfsdk:"database"`
	Roles           types.List   `tfsdk:"roles"`
	Server          *Server      `tfsdk:"server"`
}

type AadLogin struct {
	AadLoginName    types.String `tfsdk:"aad_login_name"`
	DefaultDatabase types.String `tfsdk:"default_database"`
	DefaultLanguage types.String `tfsdk:"default_language"`
	Server          *Server      `tfsdk:"server"`
}

type SqlUserLogin struct {
	LoginName       types.String `tfsdk:"login_name"`
	LoginPassword   types.String `tfsdk:"login_password"`
	DefaultDatabase types.String `tfsdk:"default_database"`
	DefaultLanguage types.String `tfsdk:"default_language"`
	Server          *Server      `tfsdk:"server"`
}

type Server struct {
	Host     types.String `tfsdk:"host"`
	Port     types.String `tfsdk:"port"`
	AzLogin  *AzLogin     `tfsdk:"azure_login"`
	SqlLogin *SqlLogin    `tfsdk:"sql_login"`
	AzureCli *AzureCli    `tfsdk:"azure_cli"`
}

type AzLogin struct {
	TenantId     types.String `tfsdk:"tenant_id"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

type SqlLogin struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type AzureCli struct {
}
