terraform {
  required_version = ">= 1.3.0"
  required_providers {
    mssql = {
      source = "softwarehuset/mssql"
    }
  }
}

provider "mssql" {
}

resource "mssql_aad_login_v2" "fun" {
  aad_login_name = "pers@bunker-holding.com"
  server = {
    #    sql_login = {
    #      username = "nico"
    #
    #      password = "Qi9PTbKwebW8XOfKZvyvfeXJu7QZq8HwG4LER7QgY7gzGnXaBQmtKdpwZ7a4XkD7"
    #    }
    "azure_cli" = {

    }
    #
    #    azure_login = {
    #      tenant_id = "fun"
    #
    #      client_id = "fun"
    #      client_secret = "fun"
    #    }
    host = "nico-3-t.database.windows.net"
  }
}


resource "mssql_sql_login_v2" "fun2" {
  login_name     = "funfun2"
  login_password = "dsauifafbg&daskda13!DA"
  server = {
    #    sql_login = {
    #      username = "nico"
    #
    #      password = "Qi9PTbKwebW8XOfKZvyvfeXJu7QZq8HwG4LER7QgY7gzGnXaBQmtKdpwZ7a4XkD7"
    #    }
    "azure_cli" = {
    }
    #
    #    azure_login = {
    #      tenant_id = "fun"
    #
    #      client_id = "fun"
    #      client_secret = "fun"
    #    }
    host = "nico-3-t.database.windows.net"
  }
}