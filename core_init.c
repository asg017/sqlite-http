#include "http.h"
int core_init(const char *dummy) {
  return sqlite3_auto_extension((void *) sqlite3_http_init);
}
