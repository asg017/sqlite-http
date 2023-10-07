.load target/debug/libsqlite_http sqlite3_http_init

select http_get_body();
