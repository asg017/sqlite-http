require "version"

module SqliteHttp
  class Error < StandardError; end
  def self.http_loadable_path
    File.expand_path('../http0', __FILE__)
  end
  def self.load(db)
    db.load_extension(self.http_loadable_path)
  end
end
