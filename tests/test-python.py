import unittest
import sqlite3
import sqlite_http

class TestSqliteHttpPython(unittest.TestCase):
  def test_path(self):
    self.assertEqual(type(sqlite_http.loadable_path()), str)
  
  def test_load(self):
    db = sqlite3.connect(':memory:')
    db.enable_load_extension(True)
    sqlite_http.load(db)
    version, = db.execute('select http_version()').fetchone()
    self.assertEqual(version[0], "v")
    
    rate_limit, = db.execute('select http_get_body("https://api.github.com/rate_limit")').fetchone()
    self.assertTrue(len(rate_limit) > 100)
  
  def test_load_no_network(self):
    db = sqlite3.connect(':memory:')
    db.enable_load_extension(True)
    sqlite_http.load_no_network(db)
    version, = db.execute('select http_version()').fetchone()
    self.assertEqual(version[0], "v")
    
    with self.assertRaisesRegex(sqlite3.OperationalError, "no such function: http_get_body"):
      db.execute('select http_get_body("https://api.github.com/rate_limit")').fetchone()

if __name__ == '__main__':
    unittest.main()