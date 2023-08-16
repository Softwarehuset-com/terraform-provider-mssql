package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/denisenkom/go-mssqldb/azuread"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os/exec"
)

type AccessTokenConnector struct {
	connection Connector
}

func (c *AccessTokenConnector) Connect(ctx context.Context) (driver.Conn, error) {
	tflog.Info(ctx, "fixing", map[string]interface{}{})
	getTokenFromCommand := exec.Command("az", "account", "get-access-token", "--scope=https://database.windows.net/.default", "--query=accessToken", "--output=tsv")
	accessToken, err := getTokenFromCommand.Output()
	tflog.Info(ctx, "generated access token", map[string]interface{}{
		"accessToken": string(accessToken),
	})
	if err != nil {
		return nil, err
	}
	tflog.Info(ctx, "finallyy CALLECED!!!")

	stringConn := fmt.Sprintf("server=%s;port=%s;password=%s;database=%s;fedauth=ActiveDirectoryServicePrincipalAccessToken;",
		c.connection.Host, c.connection.Port, string(accessToken), "master")

	db, err := sql.Open(azuread.DriverName, stringConn)
	return db.Driver().Open(stringConn)
}

func (c *AccessTokenConnector) Driver() driver.Driver {
	//TODO implement me
	panic("implement me")
}
