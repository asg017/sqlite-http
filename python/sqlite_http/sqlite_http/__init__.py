import os
import sqlite3

from sqlite_http.version import __version_info__, __version__ 

ENTRYPOINT_NO_NETWORK = "sqlite3_http_no_network_init"

def loadable_path():
  loadable_path = os.path.join(os.path.dirname(__file__), "http0")
  return os.path.normpath(loadable_path)

def load(connection: sqlite3.Connection)  -> None:
  connection.load_extension(loadable_path())

def load_no_network(connection: sqlite3.Connection)  -> None:
  connection.execute('select load_extension(?, ?)', [loadable_path(), ENTRYPOINT_NO_NETWORK]).fetchone()
