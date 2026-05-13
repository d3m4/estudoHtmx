use crate::db::Db;
use rusqlite::params;

#[derive(Debug, Clone)]
pub struct Expense {
    pub id: i64,
    pub nome: String,
    pub valor: f64,
    pub descricao: String,
    pub total_acumulado: f64,
}

pub const PAGE_SIZE: i64 = 20;

pub fn list_page(db: &Db, page: i64) -> rusqlite::Result<(Vec<Expense>, i64)> {
    let page = if page < 1 { 1 } else { page };
    let offset = (page - 1) * PAGE_SIZE;

    let conn = db.lock().unwrap();

    let total: i64 = conn.query_row("SELECT COUNT(*) FROM expenses", [], |r| r.get(0))?;

    let mut stmt = conn.prepare(
        "SELECT id, nome, valor, descricao, \
                SUM(valor) OVER (ORDER BY id) AS total_acumulado \
           FROM expenses \
          ORDER BY id \
          LIMIT ?1 OFFSET ?2",
    )?;

    let rows = stmt
        .query_map(params![PAGE_SIZE, offset], |row| {
            Ok(Expense {
                id: row.get(0)?,
                nome: row.get(1)?,
                valor: row.get(2)?,
                descricao: row.get(3)?,
                total_acumulado: row.get(4)?,
            })
        })?
        .collect::<Result<Vec<_>, _>>()?;

    Ok((rows, total))
}

pub fn get_by_id(db: &Db, id: i64) -> rusqlite::Result<Option<Expense>> {
    let conn = db.lock().unwrap();
    let res = conn.query_row(
        "SELECT id, nome, valor, descricao, 0.0 as total_acumulado \
           FROM expenses WHERE id = ?1",
        params![id],
        |row| {
            Ok(Expense {
                id: row.get(0)?,
                nome: row.get(1)?,
                valor: row.get(2)?,
                descricao: row.get(3)?,
                total_acumulado: row.get(4)?,
            })
        },
    );
    match res {
        Ok(e) => Ok(Some(e)),
        Err(rusqlite::Error::QueryReturnedNoRows) => Ok(None),
        Err(e) => Err(e),
    }
}

pub fn insert(db: &Db, nome: &str, valor: f64, descricao: &str) -> rusqlite::Result<i64> {
    let conn = db.lock().unwrap();
    conn.execute(
        "INSERT INTO expenses (nome, valor, descricao) VALUES (?1, ?2, ?3)",
        params![nome, valor, descricao],
    )?;
    Ok(conn.last_insert_rowid())
}

pub fn update(db: &Db, id: i64, nome: &str, valor: f64, descricao: &str) -> rusqlite::Result<bool> {
    let conn = db.lock().unwrap();
    let n = conn.execute(
        "UPDATE expenses SET nome = ?1, valor = ?2, descricao = ?3 WHERE id = ?4",
        params![nome, valor, descricao, id],
    )?;
    Ok(n > 0)
}

pub fn zerar(db: &Db, id: i64) -> rusqlite::Result<bool> {
    let conn = db.lock().unwrap();
    let n = conn.execute("UPDATE expenses SET valor = 0 WHERE id = ?1", params![id])?;
    Ok(n > 0)
}

pub fn sum_all(db: &Db) -> rusqlite::Result<f64> {
    let conn = db.lock().unwrap();
    conn.query_row("SELECT COALESCE(SUM(valor), 0) FROM expenses", [], |r| {
        r.get::<_, f64>(0)
    })
}

pub fn sum_excluding(db: &Db, excluding_id: i64) -> rusqlite::Result<f64> {
    let conn = db.lock().unwrap();
    conn.query_row(
        "SELECT COALESCE(SUM(valor), 0) FROM expenses WHERE id != ?1",
        params![excluding_id],
        |r| r.get::<_, f64>(0),
    )
}

pub fn total_pages(total_count: i64) -> i64 {
    if total_count <= 0 {
        1
    } else {
        (total_count + PAGE_SIZE - 1) / PAGE_SIZE
    }
}
