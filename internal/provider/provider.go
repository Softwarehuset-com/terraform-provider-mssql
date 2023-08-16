package provider

import (
	"context"
	"database/sql/driver"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/softwarehuset/mssql/internal/model"
	"github.com/softwarehuset/mssql/internal/sql"
	"net/http"
)

var (
	_ provider.Provider = &Provider{}
	_ driver.Connector  = &sql.AccessTokenConnector{}
)

type Provider struct {
	version string
}

type ProviderModel struct {
}

func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mssql"
	resp.Version = p.version
}

func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		//Attributes: map[string]schema.Attribute{
		//	"endpoint": schema.StringAttribute{
		//		MarkdownDescription: "Example provider attribute",
		//		Optional:            true,
		//	},
		//},
	}
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAadLoginResource,
		NewSqlLoginResource,
	}
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version: version,
		}
	}
}

func GetSqlClientFactory() model.SqlClientFactory {

	return sql.GetFactory()
}
