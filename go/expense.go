package main

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
