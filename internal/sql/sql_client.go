package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pkg/errors"
	"github.com/softwarehuset/mssql/internal/model"
	"net/url"
	"strings"
	"time"
)

type factory struct{}

func GetFactory() model.SqlClientFactory {
	return new(factory)
}

func (f factory) GetSqlClient(ctx context.Context, server model.Server, database string) (interface{}, error) {
	connector := &Connector{
		Host:     server.Host.ValueString(),
		Database: database,
		Port: func() string {
			if server.Port.IsNull() {
				return "1433"
			}
			return server.Port.ValueString()
		}(),
		Timeout: 180 * time.Second,
	}

	if server.SqlLogin != nil {
		tflog.Info(ctx, "Using SQL Login", map[string]interface{}{})
		connector.Login = &LoginUser{
			Username: server.SqlLogin.Username.ValueString(),
			Password: server.SqlLogin.Password.ValueString(),
		}
	}

	if server.AzureCli != nil {
		connector.AzureCli = &model.AzureCli{}
	}

	if server.AzLogin != nil {
		connector.AzureLogin = &AzureLogin{
			TenantID:     server.AzLogin.TenantId.String(),
			ClientID:     server.AzLogin.ClientId.String(),
			ClientSecret: server.AzLogin.ClientSecret.String(),
		}
	}

	return connector, nil
}

type Connector struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	Database    string `json:"database"`
	Login       *LoginUser
	AzureLogin  *AzureLogin
	AzureCli    *model.AzureCli
	AccessToken *AccessToken
	Timeout     time.Duration `json:"timeout,omitempty"`
	Token       string
}

type LoginUser struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type AzureLogin struct {
	TenantID     string `json:"tenant_id,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

type FedauthMSI struct {
	UserID string `json:"user_id,omitempty"`
}

type AccessToken struct {
	AccessToken string `json:"access_token,omitempty"`
}

func (c *Connector) PingContext(ctx context.Context) error {
	db, err := c.db(ctx)
	if err != nil {
		return err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return errors.Wrap(err, "In ping")
	}

	return nil
}

// Execute an SQL statement and ignore the results
func (c *Connector) ExecContext(ctx context.Context, command string, args ...interface{}) error {
	db, err := c.db(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, command, args...)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connector) QueryContext(ctx context.Context, query string, scanner func(*sql.Rows) error, args ...interface{}) error {
	db, err := c.db(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	err = scanner(rows)
	if err != nil {
		return err
	}

	return nil
}

func (c *Connector) QueryRowContext(ctx context.Context, query string, scanner func(*sql.Row) error, args ...interface{}) error {
	db, err := c.db(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	row := db.QueryRowContext(ctx, query, args...)
	if row.Err() != nil {
		return row.Err()
	}

	return scanner(row)
}

func (c *Connector) db(ctx context.Context) (*sql.DB, error) {
	if c == nil {
		panic("No connector")
	}
	conn, err := c.connector(ctx)
	if err != nil {
		return nil, err
	}
	if db, err := connectLoop(ctx, conn, c.Timeout); err != nil {
		return nil, err
	} else {
		return db, nil
	}
}

func (c *Connector) connector(ctx context.Context) (driver.Connector, error) {
	query := url.Values{}
	host := fmt.Sprintf("%s:%s", c.Host, c.Port)
	if c.Database != "" {
		query.Set("database", c.Database)
	}

	if c.Login != nil {
		connectionString := (&url.URL{
			Scheme:   "sqlserver",
			User:     c.userPassword(),
			Host:     host,
			RawQuery: query.Encode(),
		}).String()
		return mssql.NewConnector(connectionString)
	}
	if c.AzureCli != nil {

		return &AccessTokenConnector{connection: *c}, nil
	}

	return nil, nil
}

func (c *Connector) userPassword() *url.Userinfo {
	if c.Login != nil {
		return url.UserPassword(c.Login.Username, c.Login.Password)
	}
	return nil
}

func (c *Connector) tokenProvider() (string, error) {
	const resourceID = "https://database.windows.net/"

	admin := c.AzureLogin
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, admin.TenantID)
	if err != nil {
		return "", err
	}

	spt, err := adal.NewServicePrincipalToken(*oauthConfig, admin.ClientID, admin.ClientSecret, resourceID)
	if err != nil {
		return "", err
	}

	err = spt.EnsureFresh()
	if err != nil {
		return "", err
	}

	c.Token = spt.OAuthToken()

	return spt.OAuthToken(), nil
}

func connectLoop(ctx context.Context, connector driver.Connector, timeout time.Duration) (*sql.DB, error) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	timeoutExceeded := time.After(timeout)
	for {
		select {
		case <-timeoutExceeded:
			return nil, fmt.Errorf("db connection failed after %s timeout", timeout)

		case <-ticker.C:
			db, err := connect(ctx, connector)
			if err == nil {
				return db, nil
			}
			if strings.Contains(err.Error(), "Login failed") {
				return nil, err
			}
			if strings.Contains(err.Error(), "Login error") {
				return nil, err
			}
			if strings.Contains(err.Error(), "error retrieving access token") {
				return nil, err
			}
			tflog.Info(ctx, err.Error())
		}
	}
}

func connect(ctx context.Context, connector driver.Connector) (*sql.DB, error) {
	db := sql.OpenDB(connector)
	if err := db.Ping(); err != nil {
		tflog.Info(ctx, err.Error())

		db.Close()
		return nil, err
	}
	return db, nil
}
