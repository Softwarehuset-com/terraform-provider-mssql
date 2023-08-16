package model

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SqlClientFactory interface {
	GetSqlClient(ctx context.Context, server Server) (interface{}, error)
}

type Login struct {
	PrincipalID     int64
	LoginName       string
	DefaultDatabase string
	DefaultLanguage string
}

type User struct {
	PrincipalID     int64
	Username        string
	ObjectId        string
	LoginName       string
	Password        string
	SIDStr          string
	AuthType        string
	DefaultSchema   string
	DefaultLanguage string
	Roles           []string
}

type AadLogin struct {
	LoginName       types.String `tfsdk:"login_name"`
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
