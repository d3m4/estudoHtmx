package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// openDB abre a conexao SQLite contra shared/expenses.db (na raiz do repo,
// um nivel acima do diretorio do binario go/) e garante que o schema foi
// aplicado. Idempotente.
func openDB() (*sql.DB, error) {
	dbPath, err := filepath.Abs(filepath.Join("..", "shared", "expenses.db"))
	if err != nil {
		return nil, fmt.Errorf("resolver caminho do banco: %w", err)
	}
	schemaPath, err := filepath.Abs(filepath.Join("..", "shared", "schema.sql"))
	if err != nil {
		return nil, fmt.Errorf("resolver caminho do schema: %w", err)
	}

	// garante que o diretorio shared/ existe (no-op se ja existir)
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("criar diretorio shared: %w", err)
	}

	// DSN com cache compartilhado e foreign keys ligadas; modo WAL e ligado
	// pelo proprio schema.sql via PRAGMA.
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("abrir banco: %w", err)
	}

	// SQLite + um unico processo: 1 conexao evita "database is locked"
	// em writes concorrentes sem perder ergonomia.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping no banco: %w", err)
	}

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ler schema.sql: %w", err)
	}

	if _, err := db.Exec(string(schemaBytes)); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("aplicar schema: %w", err)
	}

	return db, nil
}
