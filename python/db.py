"""Helper de conexao sqlite3 + bootstrap de schema."""
from __future__ import annotations

import sqlite3
from pathlib import Path

# caminho absoluto pro banco compartilhado em shared/expenses.db
PYTHON_DIR = Path(__file__).resolve().parent
REPO_ROOT = PYTHON_DIR.parent
SHARED_DIR = REPO_ROOT / "shared"
DB_PATH = SHARED_DIR / "expenses.db"
SCHEMA_PATH = SHARED_DIR / "schema.sql"


def connect() -> sqlite3.Connection:
    """Abre nova conexao sqlite3 com row_factory=Row."""
    conn = sqlite3.connect(DB_PATH, isolation_level=None)  # autocommit
    conn.row_factory = sqlite3.Row
    conn.execute("PRAGMA foreign_keys = ON")
    return conn


def ensure_schema() -> None:
    """Roda shared/schema.sql (idempotente). Chamar no startup do app."""
    SHARED_DIR.mkdir(parents=True, exist_ok=True)
    schema_sql = SCHEMA_PATH.read_text(encoding="utf-8")
    with connect() as conn:
        conn.executescript(schema_sql)
