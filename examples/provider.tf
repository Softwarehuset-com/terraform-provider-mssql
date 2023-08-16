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
  login_name = "sida.epi@comp-host.com"
  server     = {
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
    host = "nico-pro-2.database.windows.net"
  }
}