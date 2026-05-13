import { db } from "./db.ts";
import { PAGE_SIZE, type Expense, type ExpenseRow } from "./types.ts";

// SELECT paginado com running sum global via window function
// SUM(valor) OVER (ORDER BY id) calcula sobre o set inteiro antes do LIMIT,
// portanto pagina 2 comeca de onde pagina 1 terminou.
const listStmt = db.prepare<[number]>(`
  SELECT id, nome, valor, descricao, created_at, total_acumulado
    FROM (
      SELECT id, nome, valor, descricao, created_at,
             SUM(valor) OVER (ORDER BY id) AS total_acumulado
        FROM expenses
    )
   ORDER BY id
   LIMIT ${PAGE_SIZE} OFFSET ?
`);

const countStmt = db.prepare(`SELECT COUNT(*) AS c FROM expenses`);
const getByIdStmt = db.prepare<[number]>(`SELECT id, nome, valor, descricao, created_at FROM expenses WHERE id = ?`);
const insertStmt = db.prepare(`INSERT INTO expenses (nome, valor, descricao) VALUES (?, ?, ?)`);
const updateStmt = db.prepare(`UPDATE expenses SET nome = ?, valor = ?, descricao = ? WHERE id = ?`);
const zerarStmt = db.prepare(`UPDATE expenses SET valor = 0 WHERE id = ?`);
const sumAllStmt = db.prepare(`SELECT COALESCE(SUM(valor), 0) AS s FROM expenses`);
const sumExcludingStmt = db.prepare<[number]>(`SELECT COALESCE(SUM(valor), 0) AS s FROM expenses WHERE id <> ?`);

export function listExpenses(page: number): { rows: ExpenseRow[]; total: number; totalPages: number; page: number } {
  const safePage = Math.max(1, Math.floor(page) || 1);
  const offset = (safePage - 1) * PAGE_SIZE;
  const rows = listStmt.all(offset) as ExpenseRow[];
  const total = (countStmt.get() as { c: number }).c;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  return { rows, total, totalPages, page: safePage };
}

export function countExpenses(): number {
  return (countStmt.get() as { c: number }).c;
}

export function getExpenseById(id: number): Expense | undefined {
  return getByIdStmt.get(id) as Expense | undefined;
}

export function insertExpense(nome: string, valor: number, descricao: string): number {
  const result = insertStmt.run(nome, valor, descricao);
  return Number(result.lastInsertRowid);
}

export function updateExpense(id: number, nome: string, valor: number, descricao: string): void {
  updateStmt.run(nome, valor, descricao, id);
}

export function zerarExpense(id: number): void {
  zerarStmt.run(id);
}

export function sumOfAll(): number {
  return (sumAllStmt.get() as { s: number }).s;
}

export function sumExcluding(id: number): number {
  return (sumExcludingStmt.get(id) as { s: number }).s;
}

// dado um id ja salvo, descobrir em que pagina ele caiu (pra POST retornar a pagina certa)
const pageOfStmt = db.prepare<[number]>(`SELECT COUNT(*) AS c FROM expenses WHERE id <= ?`);
export function pageOfId(id: number): number {
  const c = (pageOfStmt.get(id) as { c: number }).c;
  return Math.max(1, Math.ceil(c / PAGE_SIZE));
}
