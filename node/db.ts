import Database from "better-sqlite3";
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

// caminho do db compartilhado entre os 5 stacks
// resolvido a partir da pasta `node/` -> sobe um nivel -> shared/expenses.db
// usamos fileURLToPath para compatibilidade Bun e Node (sem import.meta.dir)
const __dirname = dirname(fileURLToPath(import.meta.url));
const DB_PATH = resolve(__dirname, "..", "shared", "expenses.db");
const SCHEMA_PATH = resolve(__dirname, "..", "shared", "schema.sql");

export const db = new Database(DB_PATH);

// pragmas + schema idempotente
db.pragma("journal_mode = WAL");
db.pragma("foreign_keys = ON");

const schemaSql = readFileSync(SCHEMA_PATH, "utf8");
db.exec(schemaSql);

console.log(`[db] sqlite conectado em ${DB_PATH}`);
