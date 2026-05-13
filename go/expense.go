package main

import (
	"database/sql"
	"errors"
)

// Expense representa uma despesa unica armazenada no banco.
// Os campos seguem o schema canonical (shared/schema.sql).
type Expense struct {
	Id        int64
	Nome      string
	Valor     float64
	Descricao string
	CreatedAt string
}

// ExpenseRow estende Expense com o running sum (SUM(valor) OVER (ORDER BY id)).
// Usado pela listagem paginada da tabela.
type ExpenseRow struct {
	Expense
	TotalAcumulado float64
}

// pageSize define quantos registros por pagina na listagem.
const pageSize = 20

// listExpenses retorna uma pagina de expenses (1-indexed) com o running sum
// global calculado via window function. Tambem retorna o total de registros
// para calculo de paginacao.
func listExpenses(db *sql.DB, page int) ([]ExpenseRow, int, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	total, err := countExpenses(db)
	if err != nil {
		return nil, 0, err
	}

	const q = `
SELECT id, nome, valor, descricao, created_at,
       SUM(valor) OVER (ORDER BY id) AS total_acumulado
  FROM expenses
 ORDER BY id
 LIMIT ? OFFSET ?`

	rows, err := db.Query(q, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ExpenseRow, 0, pageSize)
	for rows.Next() {
		var r ExpenseRow
		if err := rows.Scan(&r.Id, &r.Nome, &r.Valor, &r.Descricao, &r.CreatedAt, &r.TotalAcumulado); err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// countExpenses retorna a contagem total de registros.
func countExpenses(db *sql.DB) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM expenses`).Scan(&n)
	return n, err
}

// getExpense busca um registro pelo id. Retorna (nil, nil) se nao existir.
func getExpense(db *sql.DB, id int64) (*Expense, error) {
	const q = `SELECT id, nome, valor, descricao, created_at FROM expenses WHERE id = ?`
	var e Expense
	err := db.QueryRow(q, id).Scan(&e.Id, &e.Nome, &e.Valor, &e.Descricao, &e.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// insertExpense insere um novo registro e retorna o id gerado.
func insertExpense(db *sql.DB, nome string, valor float64, descricao string) (int64, error) {
	const q = `INSERT INTO expenses (nome, valor, descricao) VALUES (?, ?, ?)`
	res, err := db.Exec(q, nome, valor, descricao)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// updateExpense atualiza nome/valor/descricao de um registro existente.
func updateExpense(db *sql.DB, id int64, nome string, valor float64, descricao string) error {
	const q = `UPDATE expenses SET nome = ?, valor = ?, descricao = ? WHERE id = ?`
	_, err := db.Exec(q, nome, valor, descricao, id)
	return err
}

// zerarExpense seta valor=0 para o registro indicado.
func zerarExpense(db *sql.DB, id int64) error {
	const q = `UPDATE expenses SET valor = 0 WHERE id = ?`
	_, err := db.Exec(q, id)
	return err
}

// sumAll retorna a soma de todos os valores. Se excludingId > 0, exclui esse
// registro da soma (usado pelo total_acumulado do form em modo edicao).
func sumAll(db *sql.DB, excludingId int64) (float64, error) {
	var q string
	var args []any
	if excludingId > 0 {
		q = `SELECT COALESCE(SUM(valor), 0) FROM expenses WHERE id <> ?`
		args = []any{excludingId}
	} else {
		q = `SELECT COALESCE(SUM(valor), 0) FROM expenses`
	}
	var s float64
	if err := db.QueryRow(q, args...).Scan(&s); err != nil {
		return 0, err
	}
	return s, nil
}
