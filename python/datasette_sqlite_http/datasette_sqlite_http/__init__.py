from datasette import hookimpl
import sqlite_http

from datasette_sqlite_http.version import __version_info__, __version__ 

@hookimpl
def prepare_connection(conn, database, datasette):
    config = (
        datasette.plugin_config("datasette-sqlite-http", database=database)
        or {}
    )
    
    conn.enable_load_extension(True)
    
    if config.get("UNSAFE_allow_http_requests"):
       sqlite_http.load(conn)
    else:
      sqlite_http.load_no_network(conn)

    conn.enable_load_extension(False)