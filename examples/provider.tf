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
#
resource "mssql_aad_login_v2" "fun" {
  aad_login_name = "math.pro@bunker-holding.com"
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
    host = "nic.database.windows.net"
  }
}

#resource "mssql_aad_login_v2" "group" {
#  aad_login_name = "BOP-D-ATOZ-Owner"
#  server         = {
#    #    sql_login = {
#    #      username = "nico"
#    #
#    #      password = "Qi9PTbKwebW8XOfKZvyvfeXJu7QZq8HwG4LER7QgY7gzGnXaBQmtKdpwZ7a4XkD7"
#    #    }
#    "azure_cli" = {
#
#    }
#    #
#    #    azure_login = {
#    #      tenant_id = "fun"
#    #
#    #      client_id = "fun"
#    #      client_secret = "fun"
#    #    }
#    host = "nico-pro-3.database.windows.net"
#  }
#}


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
    host = "nic.database.windows.net"
  }
}
#
resource "mssql_sql_user_v2" "fun" {
  database   = "nic"
  username   = "nico"
  login_name = mssql_sql_login_v2.fun2.login_name
  server = {
    "azure_cli" = {
    }
    host = "nic.database.windows.net"
  }
  roles = ["db_datareader", "db_datawriter", "db_owner"]
}

#resource "mssql_sql_user_v2" "fun2" {
#  database   = "nic"
#  username   = mssql_aad_login_v2.fun.aad_login_name
#  login_name = mssql_aad_login_v2.fun.aad_login_name
#  server     = {
#    "azure_cli" = {
#    }
#    host = "nic.database.windows.net"
#  }
#  roles = ["db_owner", "db_datareader", "db_datawriter"]
#}

#resource "mssql_sql_user_v2" "group" {
#  database   = "master"
#  username = mssql_aad_login_v2.group.aad_login_name
#  login_name = mssql_aad_login_v2.group.aad_login_name
#  server         = {
#    "azure_cli" = {
#    }
#    host = "nico-pro-3.database.windows.net"  }
#  roles = ["db_owner"]
#}