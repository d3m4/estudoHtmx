use rusqlite::Connection;
use std::path::Path;
use std::sync::{Arc, Mutex};

pub type Db = Arc<Mutex<Connection>>;

pub fn open_and_ensure_schema(db_path: &str, schema_path: &str) -> rusqlite::Result<Db> {
    if let Some(parent) = Path::new(db_path).parent() {
        if !parent.as_os_str().is_empty() {
            std::fs::create_dir_all(parent).ok();
        }
    }

    let conn = Connection::open(db_path)?;

    conn.pragma_update(None, "journal_mode", "WAL")?;
    conn.pragma_update(None, "foreign_keys", "ON")?;

    let schema = std::fs::read_to_string(schema_path)
        .unwrap_or_else(|e| panic!("falha lendo schema {schema_path}: {e}"));
    conn.execute_batch(&schema)?;

    Ok(Arc::new(Mutex::new(conn)))
}
