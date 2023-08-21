package sql

import (
	"context"
	sql "database/sql"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/softwarehuset/mssql/internal/model"
	"strings"
)

var _ UserConnector = &Connector{}

func (c *Connector) GetUser(ctx context.Context, database, username string) (*model.User, error) {
	cmd := `DECLARE @stmt nvarchar(max)
          IF @@VERSION LIKE 'Microsoft SQL Azure%'
            BEGIN
              SET @stmt = 'WITH CTE_Roles (principal_id, role_principal_id) AS ' +
                          '(' +
                          '  SELECT member_principal_id, role_principal_id FROM [sys].[database_role_members] WHERE member_principal_id = DATABASE_PRINCIPAL_ID(' + QuoteName(@username, '''') + ')' +
                          '  UNION ALL ' +
                          '  SELECT member_principal_id, drm.role_principal_id FROM [sys].[database_role_members] drm' +
                          '    INNER JOIN CTE_Roles cr ON drm.member_principal_id = cr.role_principal_id' +
                          ') ' +
                          'SELECT p.principal_id, p.name, p.authentication_type_desc, COALESCE(p.default_schema_name, ''''), COALESCE(p.default_language_name, ''''), p.sid, CONVERT(VARCHAR(1000), p.sid, 1) AS sidStr, '''', COALESCE(STRING_AGG(USER_NAME(r.role_principal_id), '',''), '''') ' +
                          'FROM [sys].[database_principals] p' +
                          '  LEFT JOIN CTE_Roles r ON p.principal_id = r.principal_id ' +
                          'WHERE p.name = ' + QuoteName(@username, '''') + ' ' +
                          'GROUP BY p.principal_id, p.name, p.authentication_type_desc, p.default_schema_name, p.default_language_name, p.sid'
            END
          ELSE
            BEGIN
              SET @stmt = 'WITH CTE_Roles (principal_id, role_principal_id) AS ' +
                          '(' +
                          '  SELECT member_principal_id, role_principal_id FROM ' + QuoteName(@database) + '.[sys].[database_role_members] WHERE member_principal_id = DATABASE_PRINCIPAL_ID(' + QuoteName(@username, '''') + ')' +
                          '  UNION ALL ' +
                          '  SELECT member_principal_id, drm.role_principal_id FROM ' + QuoteName(@database) + '.[sys].[database_role_members] drm' +
                          '    INNER JOIN CTE_Roles cr ON drm.member_principal_id = cr.role_principal_id' +
                          ') ' +
                          'SELECT p.principal_id, p.name, p.authentication_type_desc, COALESCE(p.default_schema_name, ''''), COALESCE(p.default_language_name, ''''), p.sid, CONVERT(VARCHAR(1000), p.sid, 1) AS sidStr, COALESCE(sl.name, ''''), COALESCE(STRING_AGG(USER_NAME(r.role_principal_id), '',''), '''') ' +
                          'FROM ' + QuoteName(@database) + '.[sys].[database_principals] p' +
                          '  LEFT JOIN CTE_Roles r ON p.principal_id = r.principal_id ' +
                          '  LEFT JOIN [master].[sys].[sql_logins] sl ON p.sid = sl.sid ' +
                          'WHERE p.name = ' + QuoteName(@username, '''') + ' ' +
                          'GROUP BY p.principal_id, p.name, p.authentication_type_desc, p.default_schema_name, p.default_language_name, p.sid, sl.name'
            END
          EXEC (@stmt)`
	var (
		user              model.User
		sid               []byte
		roles             string
		principalId       int64
		capturedUserName  string
		capturedAuthType  string
		capturedSchema    string
		capturedLanguage  string
		capturedSidStr    string
		capturedLoginName string
	)
	err := c.
		setDatabase(ctx, &database).
		QueryRowContext(ctx, cmd,
			func(r *sql.Row) error {
				var result = r.Scan(&principalId, &capturedUserName, &capturedAuthType, &capturedSchema, &capturedLanguage, &sid, &capturedSidStr, &capturedLoginName, &roles)
				user.PrincipalID = types.Int64Value(principalId)

				user.AuthType = types.StringValue(capturedAuthType)
				user.UserName = types.StringValue(capturedUserName)
				user.DefaultSchema = types.StringValue(capturedSchema)
				user.DefaultLanguage = types.StringValue(capturedLanguage)
				user.SIDStr = types.StringValue(capturedSidStr)
				user.LoginName = types.StringValue(capturedLoginName)

				return result
			},
			sql.Named("database", database),
			sql.Named("username", username),
		)
	if err != nil {
		tflog.Warn(ctx, "No servicePrincipal rows were returned.")
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	tflog.Info(ctx, "User: "+user.LoginName.ValueString())
	if user.AuthType.ValueString() == "INSTANCE" && capturedLoginName == "" {
		cmd = "SELECT name FROM [master].[sys].[sql_logins] WHERE sid = @sid"
		c.Database = "master"
		err = c.QueryRowContext(ctx, cmd,
			func(r *sql.Row) error {
				return r.Scan(&capturedLoginName)
			},
			sql.Named("sid", sid),
		)
		if err != nil {
			tflog.Warn(ctx, "No sql_logins rows were returned.")

			return nil, err
		}
	}
	if roles == "" {
		user.Roles = types.List{}
	} else {
		user.Roles, _ = types.ListValueFrom(ctx, types.StringType, strings.Split(roles, ","))
	}
	return &user, nil
}

func (c *Connector) CreateUser(ctx context.Context, database string, user *model.User) error {

	var dbfun = "nic"
	// Setting Pre-reqs
	var db, err = c.setDatabase(ctx, &dbfun).db(ctx)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	tflog.Info(ctx, "EnsureSplitStringExists...")
	tflog.Info(ctx, "Using db "+database)
	err = EnsureSplitStringExists(tx, ctx)
	if err != nil {
		return err
	}
	//var existingUser *model.User
	//existingUser, err = c.GetUser(ctx, database, user.UserName.ValueString())
	//if err != nil {
	//	return err
	//}
	//if existingUser != nil {
	//	return errors.New("user already exists")
	//}
	//err = tx.Commit()
	if err != nil {
		return err
	}

	// Creating User + roles
	//tx, err = db.BeginTx(ctx, &sql.TxOptions{})

	if user.AuthType.ValueString() == "EXTERNAL" {
		tflog.Info(ctx, "Creating User as External...")

		err := CreateUserFromExternal(ctx, database, user, tx)
		if err != nil {
			return err
		}
	}

	if user.AuthType.ValueString() == "INSTANCE" {
		tflog.Info(ctx, "Creating User as INSTANCE...")

		err2 := CreateUserFromInstance(ctx, user, tx)
		if err2 != nil {
			return err2
		}
	}
	//err = tx.Commit()

	//existingUser, err = c.GetUser(ctx, database, user.UserName.ValueString())
	//if err != nil {
	//	return err
	//}

	//if existingUser == nil {
	//	return errors.New("user could not be created")
	//}
	//tx, err = db.BeginTx(ctx, &sql.TxOptions{})

	tflog.Info(ctx, "assigning roles")

	err = tx.Commit()
	if err != nil {
		return err
	}
	err = AssignRoles(ctx, user, db)
	if err != nil {
		tflog.Info(ctx, "assign roles failed")
		err2 := c.DeleteUser(ctx, database, user.UserName.ValueString())
		if err2 != nil {
			tflog.Error(ctx, "Roll back failed")
			return err
		}
		return err
	}

	tflog.Info(ctx, "committing...")

	if err != nil {
		return err
	}
	return nil
}

func AssignRoles(ctx context.Context, user *model.User, db *sql.DB) error {
	var cmd = `
		DECLARE @role nvarchar(max);
		DECLARE del_role_cur CURSOR FOR 
			select rp.name FROM [sys].[database_principals] as p
			join [sys].[database_role_members] as rm on p.principal_id = rm.member_principal_id
			join [sys].[database_principals] as rp on rm.role_principal_id = rp.principal_id 
			where 
				p.name = @user
				and rp.name not in (select * from String_Split(@roles,','))
		OPEN del_role_cur;
		FETCH NEXT FROM del_role_cur INTO @role;
		WHILE @@FETCH_STATUS = 0
			BEGIN
				PRINT 'dropping '+ @role
				EXEC sp_droprolemember @role, @user
				FETCH NEXT FROM del_role_cur INTO @role;
			END
		CLOSE del_role_cur;
		DEALLOCATE del_role_cur;
		
		DECLARE add_role_cur CURSOR FOR 
			select * from String_Split(@roles,',') where [value] not in (
				select rp.name FROM [sys].[database_principals] as p
				join [sys].[database_role_members] as rm on p.principal_id = rm.member_principal_id
				join [sys].[database_principals] as rp on rm.role_principal_id = rp.principal_id 
				where 
					p.name = @user)
		OPEN add_role_cur;
		FETCH NEXT FROM add_role_cur INTO @role;
		WHILE @@FETCH_STATUS = 0
			BEGIN
                IF @role != ''
                BEGIN
                    PRINT 'adding '+ @role
                    EXEC sp_addrolemember @role, @user
                END
                FETCH NEXT FROM add_role_cur INTO @role;
			END
		CLOSE add_role_cur;
		DEALLOCATE add_role_cur;
		`
	var rolesSlice []string
	for _, role := range user.Roles.Elements() {
		value, err := role.ToTerraformValue(ctx)
		if err != nil {
			return err
		}

		var valueString string

		err = value.As(&valueString)
		if err != nil {
			return err
		}
		rolesSlice = append(rolesSlice, valueString) // Assuming attr.Value has a .String() method
	}
	var rolesString = strings.Join(rolesSlice, ",")

	tflog.Info(ctx, "Assigning roles"+rolesString)
	_, err := db.ExecContext(ctx, cmd,
		sql.Named("user", user.UserName.ValueString()),
		sql.Named("roles", rolesString))
	if err != nil {
		return err
	}
	return nil
}

func EnsureSplitStringExists(c *sql.Tx, ctx context.Context) error {
	var cmd = `
		EXEC sp_getapplock @Resource = 'create_func', @LockMode = 'Exclusive';
	    IF exists (select compatibility_level FROM sys.databases where name = db_name() and compatibility_level < 130) AND objectproperty(object_id('String_Split'), 'isProcedure') IS NULL
	    BEGIN
	        DECLARE @sql NVARCHAR(MAX);
	        SET @sql = N'Create FUNCTION [dbo].[String_Split]
	              (
	                  @string    nvarchar(max),
	                  @delimiter nvarchar(max)
	              )
	              /*
	                  The same as STRING_SPLIT for compatibility level < 130
	                  https://docs.microsoft.com/en-us/sql/t-sql/functions/string-split-transact-sql?view=sql-server-ver15
	              */
	              RETURNS TABLE AS RETURN
	              (
	                  SELECT
	                    --ROW_NUMBER ( ) over(order by (select 0))                            AS id     --  intuitive, but not correect
	                      Split.a.value(''let $n := . return count(../*[. << $n]) + 1'', ''int'') AS id
	                    , Split.a.value(''.'', ''NVARCHAR(MAX)'')                                 AS value
	                  FROM
	                  (
	                      SELECT CAST(''<X>''+REPLACE(@string, @delimiter, ''</X><X>'')+''</X>'' AS XML) AS String
	                  ) AS a
	                  CROSS APPLY String.nodes(''/X'') AS Split(a)
	              )';
	        EXEC (@sql)
			EXEC sp_releaseapplock @Resource = 'create_func';

	     END
       `
	var _, err = c.ExecContext(ctx, cmd)
	if err != nil {
		return err
	}
	return nil
}

func CreateUserFromInstance(ctx context.Context, user *model.User, tx *sql.Tx) error {
	var query = `

				DECLARE @stmt nvarchar(max)
		SET @stmt = ''		        
	          SET @stmt = 'CREATE USER ' + QuoteName(@username) + ' FOR LOGIN ' + QuoteName(@loginName) + ' ' +
	                      'WITH DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema)

			EXEC (@stmt)`
	tflog.Info(ctx, query)

	_, err := tx.ExecContext(ctx, query,

		sql.Named("username", user.UserName.ValueString()),
		sql.Named("objectId", user.ObjectId.ValueString()),
		sql.Named("loginName", user.LoginName.ValueString()),
		sql.Named("password", user.Password.ValueString()),
		sql.Named("authType", user.AuthType.ValueString()),
		sql.Named("defaultSchema", "dbo"),
		sql.Named("defaultLanguage", user.DefaultLanguage.ValueString()),
	)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return nil
}

func CreateUserFromExternal(ctx context.Context, database string, user *model.User, tx *sql.Tx) error {

	tx.ExecContext(ctx, `
				DECLARE @stmt nvarchar(max)
				DECLARE @language nvarchar(max) = @defaultLanguage
				IF @@VERSION LIKE 'Microsoft SQL Azure%'
                BEGIN
                  IF @objectId != ''
                    BEGIN
                      SET @stmt = 'CREATE USER ' + QuoteName(@username) + ' WITH SID=' + CONVERT(varchar(64), CAST(CAST(@objectId AS UNIQUEIDENTIFIER) AS VARBINARY(16)), 1) + ', TYPE=E'
                    END
                  ELSE
                    BEGIN
                      SET @stmt = 'CREATE USER ' + QuoteName(@username) + ' FROM EXTERNAL PROVIDER'
                  END
                END
				ELSE
					BEGIN
					  SET @stmt = 'CREATE USER ' + QuoteName(@username) + ' FOR LOGIN ' + QuoteName(@username) + ' FROM EXTERNAL PROVIDER ' +
								  'WITH DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema) + ', ' +
								  'DEFAULT_LANGUAGE = ' + Coalesce(QuoteName(@language), 'NONE')
				END
			EXEC (@stmt)`,
		sql.Named("database", database),
		sql.Named("username", user.UserName.ValueString()),
		sql.Named("objectId", user.ObjectId.ValueString()),
		sql.Named("loginName", user.LoginName.ValueString()),
		sql.Named("password", user.Password.ValueString()),
		sql.Named("authType", user.AuthType.ValueString()),
		sql.Named("defaultSchema", user.DefaultSchema.ValueString()),
		sql.Named("defaultLanguage", user.DefaultLanguage.ValueString()),
	)
	return nil
}

func (c *Connector) UpdateUser(ctx context.Context, database string, user *model.User) error {
	cmd := `DECLARE @stmt nvarchar(max)
          SET @stmt = 'ALTER USER ' + QuoteName(@username) + ' '
          DECLARE @language nvarchar(max) = @defaultLanguage
          IF @language = '' SET @language = NULL
          SET @stmt = @stmt + 'WITH DEFAULT_SCHEMA = ' + QuoteName(@defaultSchema)
          DECLARE @auth_type nvarchar(max) = (SELECT authentication_type_desc FROM [sys].[database_principals] WHERE name = @username)
          IF NOT @@VERSION LIKE 'Microsoft SQL Azure%' AND @auth_type != 'INSTANCE'
            BEGIN
              SET @stmt = @stmt + ', DEFAULT_LANGUAGE = ' + Coalesce(QuoteName(@language), 'NONE')
            END

          BEGIN TRANSACTION;
          EXEC sp_getapplock @Resource = 'create_func', @LockMode = 'Exclusive';
          IF exists (select compatibility_level FROM sys.databases where name = db_name() and compatibility_level < 130) AND objectproperty(object_id('String_Split'), 'isProcedure') IS NULL
          BEGIN
              DECLARE @sql NVARCHAR(MAX);
              SET @sql = N'Create FUNCTION [dbo].[String_Split]
                    (
                        @string    nvarchar(max),
                        @delimiter nvarchar(max)
                    )
                    /*
                        The same as STRING_SPLIT for compatibility level < 130
                        https://docs.microsoft.com/en-us/sql/t-sql/functions/string-split-transact-sql?view=sql-server-ver15
                    */
                    RETURNS TABLE AS RETURN
                    (
                        SELECT
                          --ROW_NUMBER ( ) over(order by (select 0))                            AS id     --  intuitive, but not correect
                            Split.a.value(''let $n := . return count(../*[. << $n]) + 1'', ''int'') AS id
                          , Split.a.value(''.'', ''NVARCHAR(MAX)'')                                 AS value
                        FROM
                        (
                            SELECT CAST(''<X>''+REPLACE(@string, @delimiter, ''</X><X>'')+''</X>'' AS XML) AS String
                        ) AS a
                        CROSS APPLY String.nodes(''/X'') AS Split(a)
                    )';
              EXEC sp_executesql @sql;
          END
          EXEC sp_releaseapplock @Resource = 'create_func';
          COMMIT TRANSACTION;
          SET @stmt = @stmt + '; ' +
                      'DECLARE @sql nvarchar(max);' +
                      'DECLARE @role nvarchar(max);' +
                      'DECLARE del_role_cur CURSOR FOR SELECT name FROM ' + QuoteName(@database) + '.[sys].[database_principals] WHERE type = ''R'' AND name != ''public'' AND name IN (SELECT name FROM ' + QuoteName(@database) + '.[sys].[database_role_members] drm, ' + QuoteName(@database) + '.[sys].[database_principals] db WHERE drm.member_principal_id = DATABASE_PRINCIPAL_ID(' + QuoteName(@username, '''') + ') AND drm.role_principal_id = db.principal_id) AND name COLLATE SQL_Latin1_General_CP1_CI_AS NOT IN(SELECT value FROM String_Split(' + QuoteName(@roles, '''') + ', '',''));' +
                      'DECLARE add_role_cur CURSOR FOR SELECT name FROM ' + QuoteName(@database) + '.[sys].[database_principals] WHERE type = ''R'' AND name != ''public'' AND name NOT IN (SELECT name FROM ' + QuoteName(@database) + '.[sys].[database_role_members] drm, ' + QuoteName(@database) + '.[sys].[database_principals] db WHERE drm.member_principal_id = DATABASE_PRINCIPAL_ID(' + QuoteName(@username, '''') + ') AND drm.role_principal_id = db.principal_id) AND name COLLATE SQL_Latin1_General_CP1_CI_AS IN(SELECT value FROM String_Split(' + QuoteName(@roles, '''') + ', '',''));' +
                      'OPEN del_role_cur;' +
                      'FETCH NEXT FROM del_role_cur INTO @role;' +
                      'WHILE @@FETCH_STATUS = 0' +
                      '  BEGIN' +
                      '    SET @sql = ''ALTER ROLE '' + QuoteName(@role) + '' DROP MEMBER ' + QuoteName(@username) + ''';' +
                      '    EXEC (@sql);' +
                      '    FETCH NEXT FROM del_role_cur INTO @role;' +
                      '  END;' +
                      'CLOSE del_role_cur;' +
                      'DEALLOCATE del_role_cur;' +
                      'OPEN add_role_cur;' +
                      'FETCH NEXT FROM add_role_cur INTO @role;' +
                      'WHILE @@FETCH_STATUS = 0' +
                      '  BEGIN' +
                      '    SET @sql = ''ALTER ROLE '' + QuoteName(@role) + '' ADD MEMBER ' + QuoteName(@username) + ''';' +
                      '    EXEC (@sql);' +
                      '    FETCH NEXT FROM add_role_cur INTO @role;' +
                      '  END;' +
                      'CLOSE add_role_cur;' +
                      'DEALLOCATE add_role_cur;'
          EXEC (@stmt)`

	var rolesSlice []string
	for _, role := range user.Roles.Elements() {
		rolesSlice = append(rolesSlice, role.String()) // Assuming attr.Value has a .String() method
	}

	rolesString := strings.Join(rolesSlice, ",")

	return c.
		setDatabase(ctx, &database).
		ExecContext(ctx, cmd,
			sql.Named("database", database),
			sql.Named("username", user.UserName.ValueString()),
			sql.Named("defaultSchema", user.DefaultSchema.ValueString()),
			sql.Named("defaultLanguage", user.DefaultLanguage.ValueString()),
			sql.Named("roles", rolesString),
		)
}

func (c *Connector) DeleteUser(ctx context.Context, database, username string) error {
	cmd := `DECLARE  @stmt nvarchar(max)
          SET @stmt = 'IF EXISTS (SELECT 1 FROM [sys].[database_principals] WHERE [name] = ' + QuoteName(@username, '''') + ') ' +
                      'DROP USER ' + QuoteName(@username)
          EXEC (@stmt)`
	return c.
		setDatabase(ctx, &database).
		ExecContext(ctx, cmd, sql.Named("database", database), sql.Named("username", username))
}

func (c *Connector) setDatabase(ctx context.Context, database *string) *Connector {
	if *database == "" {
		*database = "master"
	}
	c.Database = *database

	tflog.Info(ctx, "Using database: "+c.Database)
	return c
}

type UserConnector interface {
	CreateUser(ctx context.Context, database string, user *model.User) error
	GetUser(ctx context.Context, database, username string) (*model.User, error)
	UpdateUser(ctx context.Context, database string, user *model.User) error
	DeleteUser(ctx context.Context, database, username string) error
}
