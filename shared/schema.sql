-- estudoHtmx: schema canonical compartilhado pelos 5 stacks
-- idempotente: pode rodar varias vezes sem efeito colateral

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS expenses (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    nome        TEXT    NOT NULL CHECK(length(trim(nome)) > 0),
    valor       REAL    NOT NULL DEFAULT 0,
    descricao   TEXT    NOT NULL DEFAULT '',
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_expenses_id ON expenses(id);
