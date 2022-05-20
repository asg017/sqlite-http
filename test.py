import sqlite3
import unittest
import json
import os
from datetime import datetime, timedelta

EXT_PATH = "dist/http0"
EXT_PATHNO_NET = "dist/http0-no-net"

should_skip_net = os.environ.get("SKIP_NET") == "1"

if should_skip_net:
  print("WARNING: Skipping all tests that depend on a local httpbin running. should only be used for CI tests only.")

def skip_do(f):
  def wrapper(self):
    if should_skip_net:
      self.skipTest("Skipping all do methods")
    f(self)
  return wrapper


def connect(entry:str) -> sqlite3.Cursor:
  db = sqlite3.connect(":memory:")

  db.execute("create table fbefore as select name from pragma_function_list")
  db.execute("create table mbefore as select name from pragma_module_list")

  db.enable_load_extension(True)
  db.load_extension(entry)

  db.execute("create temp table fafter as select name from pragma_function_list")
  db.execute("create temp table mafter as select name from pragma_module_list")
  
  db.row_factory = sqlite3.Row
  return db

db = connect(EXT_PATH)
db_nonet = connect(EXT_PATHNO_NET)

# Fun fact: the SQLite datetime() format, with fractional seconds,
# doesn't always have 3 digits of precision.
# so right pad timestamp with 000's until it does
# https://www.sqlite.org/lang_datefunc.html
def read_sqlite_timestamp(ts):
  ts_padded = ts.ljust(len("YYYY-MM-DD HH:MM:SS.FFF"), "0")
  return datetime.fromisoformat(ts_padded)

class TestHttp(unittest.TestCase):
  def test_funcs(self):
    funcs = list(map(lambda a: a[0], db.execute("select name from fafter where name not in (select name from fbefore) order by name").fetchall()))
    self.assertEqual(funcs, [
      "http_cookies",
      "http_debug",
      "http_do_body",
      "http_do_headers",
      "http_get_body",
      "http_get_headers",
      "http_headers",
      "http_headers_all",
      "http_headers_get",
      "http_headers_has",
      "http_post_body",
      "http_post_form_urlencoded",
      "http_post_headers",
      "http_rate_limit",
      "http_timeout_set",
      "http_version"
    ])
  
  def test_modules(self):
    funcs = list(map(lambda a: a[0], db.execute("select name from mafter where name not in (select name from mbefore) order by name").fetchall()))
    self.assertEqual(funcs, [
      "http_do",
      "http_get",
      "http_headers_each",
      "http_post",
    ])
  
  def test_nodofuncs(self):
    funcs = list(map(lambda a: a[0], db_nonet.execute("select name from fafter where name not in (select name from fbefore) order by name").fetchall()))
    self.assertEqual(funcs, [
      "http_cookies",
      "http_debug",
      "http_headers",
      "http_headers_all",
      "http_headers_get",
      "http_headers_has",
      # TODO should be a part of nodo
      #"http_post_form_urlencoded",
      "http_version"
    ])

  def test_version(self):
    v, = db.execute("select http_version()").fetchone()
    self.assertEqual(v, "v0.0.0")
  
  def test_debug(self):
    d, = db.execute("select http_debug()").fetchone()
    lines = d.splitlines()
    self.assertEqual(len(lines), 4)
    self.assertTrue(lines[0].startswith("Version"))
    self.assertTrue(lines[1].startswith("Commit"))
    self.assertTrue(lines[2].startswith("Runtime"))
    self.assertTrue(lines[3].startswith("Date"))

  def test_http_cookies(self):
    d, = db.execute("""
      select http_cookies("name", "Alex")
    """).fetchone()
    self.assertEqual(d, "{\"name\":\"Alex\"}")
  
  @skip_do
  def test_http_do_body(self):
    d, = db.execute("""
      select http_do_body(
        'DELETE', 
        'http://localhost:8080/delete'
      )
    """).fetchone()
    data = json.loads(d.decode('utf8'))
    self.assertEqual(data.get("url"), "http://localhost:8080/delete")
  
  @skip_do
  def test_http_do_headers(self):
    headers, = db.execute("""
      select http_do_headers(
        'DELETE', 
        'http://localhost:8080/delete'
      )
    """).fetchone()
    self.assertEqual(len(headers.splitlines()), 7)
  
  # TODO finish this test, then add http_post and http_do
  @skip_do
  def test_http_get(self):
    d = db.execute("""
      select * from http_get(
        'http://localhost:8080/get?name=alex',
        null, 
        http_cookies("donald", "duck")
      )
    """).fetchone()
    response_body = json.loads(d["response_body"].decode("utf8"))
    timings = json.loads(d["response_body"].decode("utf8"))
    self.assertEqual(d["request_url"], "http://localhost:8080/get?name=alex")
    self.assertEqual(d["request_method"], "GET")
    # TODO should include user-agent here
    self.assertEqual(d["request_headers"], "Cookie: donald=duck\r\n")
    # TODO this dont work lol
    #self.assertEqual(d["request_cookies"], "[]")
    self.assertEqual(d["request_body"].decode("utf8"), "")
    self.assertEqual(d["response_status"], "200 OK")
    self.assertEqual(d["response_status_code"], 200)
    self.assertEqual(len(d["response_headers"].splitlines()), 7)
    self.assertEqual(d["response_cookies"], "[]")
    self.assertTrue(len(d["response_body"]) > 100)
    self.assertTrue(d["remote_address"] in ("127.0.0.1:8080", "[::1]:8080"))
    self.assertEqual(d["meta"], None)
  
  @skip_do
  def test_http_get_body(self):
    d, = db.execute("""
      select http_get_body(
        'http://localhost:8080/get?name=alex'
      )
    """).fetchone()
    data = json.loads(d.decode("utf8"))
    self.assertEqual(data.get("args").get("name"), "alex")
  
  @skip_do
  def test_http_get_headers(self):
    headers, = db.execute("""
      select http_get_headers(
        'http://localhost:8080/get'
      )
    """).fetchone()
    self.assertEqual(len(headers.splitlines()), 7)
  
  def test_http_headers(self):
    h1, h2 = db.execute("""
      select http_headers("a", "b"), http_headers("dup", "a", "dup", "b")
    """).fetchone()
    self.assertEqual(h1, "A: b\r\n")
    self.assertEqual(h2, "Dup: a\r\nDup: b\r\n")
  
  def test_http_headers_each(self):
    rows = db.execute("""
      select * 
      from http_headers_each(
        http_headers(
          "a", "1",
          "a", "2",
          "user-agent", "4"
        )
      )
    """).fetchall()
    self.assertEqual(len(rows), 3)
    self.assertEqual(rows[0]["key"], "A")
    self.assertEqual(rows[0]["value"], "1")
    self.assertEqual(rows[1]["key"], "A")
    self.assertEqual(rows[1]["value"], "2")
    self.assertEqual(rows[2]["key"], "User-Agent")
    self.assertEqual(rows[2]["value"], "4")
    
  # TODO manually check skip_do for http_get to tests headers func seperately
  @skip_do
  def test_http_headers_get(self):
    a,b,c, d = db.execute("""
      select http_headers_get(
        http_headers("a", "xyz"), "a"
      ),
      -- gets first headers if dup
      http_headers_get(
        http_headers("a", "xyz", "a", "abc"), "a"
      ),
      -- null if not exists
      http_headers_get(
        http_headers("a", "xyz"), "b"
      ),
      http_headers_get(
        x.request_headers, 
        "X-Powered-By"
      ) 
      from http_get(
        "http://localhost:8080/get", 
        http_headers("X-Powered-By", "dogs")
      ) as x
    """).fetchone()
    self.assertEqual(a, "xyz")
    self.assertEqual(b, "xyz")
    self.assertEqual(c, None)
    self.assertEqual(d, "dogs")
  
  def test_http_headers_has(self):
    a,b = db.execute("""
      select http_headers_has(
        http_headers("a", "1"), "a"
      ),
      http_headers_has(
        http_headers("a", "1"), "b"
      )
    """).fetchone()
    self.assertEqual(a, 1)
    self.assertEqual(b, 0)
  
  @skip_do
  def test_http_post_body(self):
    d, = db.execute("""
      select http_post_body(
        "http://localhost:8080/post", 
        http_headers("content-type", "application/json"),
        json_object("name", "Alex")
      )
    """).fetchone()
    data = json.loads(d.decode("utf8"))
    self.assertEqual(data.get("json").get("name"), "Alex")
  
  # TODO test without needing to post_body
  @skip_do
  def test_http_post_form_urlencoded(self):
    d, = db.execute("""
      select http_post_body(
        'http://localhost:8080/post', 
        http_headers('Content-Type', 'application/x-www-form-urlencoded'),
        http_post_form_urlencoded(
          'name', 'Alex',
          'age', 99
        )
      )
    """).fetchone()
    data = json.loads(d.decode("utf8"))
    self.assertEqual(data.get("form").get("name"), "Alex")
    self.assertEqual(data.get("form").get("age"), "99")
  
  @skip_do
  def test_http_post_headers(self):
    headers, = db.execute("""
      select http_post_headers("http://localhost:8080/post")
    """).fetchone()
    self.assertEqual(len(headers.splitlines()), 7)
  
  # return only i, timings for testing rate_limit
  def _run_do_n(self, n=10):
    return db.execute("""
      with recursive
        count(i) as (values(1) union all select i + 1 FROM count where i < ?),
      reqs as (
        select i, timings
        from count, http_get(printf("http://localhost:8080/get?i=%s"), i)
      )
      select i, timings as curr, lag(timings) over (order by i) as prev
      from reqs
    """, (n,)).fetchall()
    
  
  @skip_do
  def test_http_rate_limit(self):
    # turn off rate limit
    db.execute("select http_rate_limit(1);")
    reqs= self._run_do_n()
    self.assertEqual(len(reqs), 10)
    
    # skip first, no prev bc of window lag
    for req in reqs[1:]:
        curr_start = read_sqlite_timestamp(json.loads(req["curr"]).get("start"))
        prev_start = read_sqlite_timestamp(json.loads(req["prev"]).get("start"))
        self.assertLess((curr_start - prev_start), timedelta(milliseconds=15))
    
    db.execute("select http_rate_limit(100);")
    reqs= self._run_do_n()
    self.assertEqual(len(reqs), 10)
    
    # skip first, no prev bc of window lag
    for req in reqs[1:]:
        curr_start = read_sqlite_timestamp(json.loads(req["curr"]).get("start"))
        prev_start = read_sqlite_timestamp(json.loads(req["prev"]).get("start"))
        self.assertGreaterEqual((curr_start - prev_start), timedelta(milliseconds=100-5))
        self.assertLessEqual((curr_start - prev_start), timedelta(milliseconds=100+5))
    
    db.execute("select http_rate_limit(20);")
    reqs= self._run_do_n()
    self.assertEqual(len(reqs), 10)
    
    # skip first, no prev bc of window lag
    for req in reqs[1:]:
        curr_start = read_sqlite_timestamp(json.loads(req["curr"]).get("start"))
        prev_start = read_sqlite_timestamp(json.loads(req["prev"]).get("start"))
        self.assertGreaterEqual((curr_start - prev_start), timedelta(milliseconds=20-3))
        self.assertLessEqual((curr_start - prev_start), timedelta(milliseconds=20+5))
        
  @skip_do
  def test_http_timeout_set(self):
    d, = db.execute("select http_timeout_set(100)").fetchone()
    self.assertEqual(d, 100)
    
    p1, = db.execute("select response_status from http_get('http://localhost:8080/delay/.05')").fetchone()
    self.assertEqual(p1, "200 OK")
    
    with self.assertRaises(sqlite3.OperationalError):
      db.execute("select response_status from http_get('http://localhost:8080/delay/2')").fetchone()
    
    d, = db.execute("select http_timeout_set(500)").fetchone()
    p3, = db.execute("select response_status from http_get('http://localhost:8080/delay/.2')").fetchone()
    self.assertEqual(p3, "200 OK")
    

if __name__ == '__main__':
    unittest.main()