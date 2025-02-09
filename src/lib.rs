use reqwest::blocking::Response;
use sqlite_loadable::prelude::*;
use sqlite_loadable::{api, define_scalar_function, Result};

pub fn http_version(context: *mut sqlite3_context, _values: &[*mut sqlite3_value]) -> Result<()> {
    api::result_text(context, format!("v{}", env!("CARGO_PKG_VERSION")))?;
    Ok(())
}

pub fn http_debug(context: *mut sqlite3_context, _values: &[*mut sqlite3_value]) -> Result<()> {
    api::result_text(
        context,
        format!(
            "Version: v{}
Source: {}
",
            env!("CARGO_PKG_VERSION"),
            env!("GIT_HASH")
        ),
    )?;
    Ok(())
}

fn has_text_content_type(response:&Response)->bool {
  let mimetype = match response.headers().get("content-type") {
    Some(m) => m.as_bytes(), 
    None => return false
  };
  mimetype.starts_with(b"text/") || mimetype.starts_with(b"application/json")
}

fn result_response(context: *mut sqlite3_context, response: Response)->Result<()> {
  if has_text_content_type(&response) {
    api::result_text(context, response.text().unwrap())?;
  }else {
    api::result_blob(context, response.bytes().unwrap().as_ref());
  }
  Ok(())
}
pub fn http_get_body(context: *mut sqlite3_context, values: &[*mut sqlite3_value]) -> Result<()> {
    let url = api::value_text(&values[0]).unwrap();
    let _headers = values.get(1).map(|v| api::value_text(v).unwrap());
    let _cookies = "";
    let client = reqwest::blocking::ClientBuilder::new().user_agent(concat!(
      env!("CARGO_PKG_NAME"),
      "/",
      env!("CARGO_PKG_VERSION"),
  )).build().unwrap();
    let request = client.get(url);
    result_response(context, request.send().unwrap())?;
    Ok(())
}
use sqlite_reader::{SqliteReader, READER_POINTER_NAME};

pub fn http_request(context: *mut sqlite3_context, values: &[*mut sqlite3_value]) -> Result<()> {
    let url = api::value_text(values.get(0).unwrap()).unwrap();

    api::result_pointer(
        context,
        READER_POINTER_NAME,
        Box::new(RequestReader {
            url: url.to_owned(),
        }) as Box<dyn SqliteReader>,
    );
    Ok(())
}

#[repr(C)]
pub struct RequestReader {
    url: String,
}
impl SqliteReader for RequestReader {
    fn generate(
        &self,
    ) -> std::result::Result<Box<dyn std::io::Read + 'static>, Box<dyn std::error::Error>> {
        let client = reqwest::blocking::Client::new();
        let x = client.get(&self.url).send().unwrap();
        Ok(Box::new(x))
    }
}

#[sqlite_entrypoint]
pub fn sqlite3_http_init(db: *mut sqlite3) -> Result<()> {
    define_scalar_function(
        db,
        "http_version",
        0,
        http_version,
        FunctionFlags::UTF8 | FunctionFlags::DETERMINISTIC,
    )?;
    define_scalar_function(
        db,
        "http_debug",
        0,
        http_debug,
        FunctionFlags::UTF8 | FunctionFlags::DETERMINISTIC,
    )?;

    define_scalar_function(db, "http_get_body", 1, http_get_body, FunctionFlags::UTF8)?;

    define_scalar_function(db, "http_request", 1, http_request, FunctionFlags::UTF8)?;
    Ok(())
}
