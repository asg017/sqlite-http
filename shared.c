// sqlite3ext.h contains the SQLite3 extension entry-point routine as defined here https://sqlite.org/loadext.html
#include "sqlite3ext.h"

SQLITE_EXTENSION_INIT3

// hook to call into golang functionality defined in go.riyazali.net/sqlite
extern int go_sqlite3_extension_init(const char*, sqlite3*, char**);

#ifdef _WIN32
__declspec(dllexport)
#endif
int sqlite3_http_init(sqlite3* db, char** pzErrMsg, const sqlite3_api_routines *pApi) {
	SQLITE_EXTENSION_INIT2(pApi)
	return go_sqlite3_extension_init("http", db, pzErrMsg);
}

#ifdef _WIN32
__declspec(dllexport)
#endif
int sqlite3_http_no_network_init(sqlite3* db, char** pzErrMsg, const sqlite3_api_routines *pApi) {
	SQLITE_EXTENSION_INIT2(pApi)
	return go_sqlite3_extension_init("http_no_network", db, pzErrMsg);
}