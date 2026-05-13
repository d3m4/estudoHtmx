"""Operacoes de acesso a dados (expenses)."""
from __future__ import annotations

from typing import Optional

from db import connect
from models import Expense

PAGE_SIZE = 20


def count() -> int:
    """Total de despesas."""
    with connect() as conn:
        row = conn.execute("SELECT COUNT(*) AS c FROM expenses").fetchone()
        return int(row["c"])


def list_page(page: int) -> list[Expense]:
    """Lista a pagina N (1-indexed), 20 itens, com running sum global."""
    if page < 1:
        page = 1
    offset = (page - 1) * PAGE_SIZE
    sql = """
        SELECT id, nome, valor, descricao, created_at,
               SUM(valor) OVER (ORDER BY id) AS total_acumulado
          FROM expenses
         ORDER BY id
         LIMIT ? OFFSET ?
    """
    with connect() as conn:
        rows = conn.execute(sql, (PAGE_SIZE, offset)).fetchall()
        return [Expense.from_row(r) for r in rows]


def get_by_id(expense_id: int) -> Optional[Expense]:
    """Pega uma despesa pelo id (sem total_acumulado)."""
    with connect() as conn:
        row = conn.execute(
            "SELECT id, nome, valor, descricao, created_at FROM expenses WHERE id = ?",
            (expense_id,),
        ).fetchone()
        if row is None:
            return None
        return Expense.from_row(row)


def insert(nome: str, valor: float, descricao: str) -> int:
    """Insere nova despesa, retorna o id criado."""
    with connect() as conn:
        cur = conn.execute(
            "INSERT INTO expenses (nome, valor, descricao) VALUES (?, ?, ?)",
            (nome, valor, descricao),
        )
        return int(cur.lastrowid)


def update(expense_id: int, nome: str, valor: float, descricao: str) -> bool:
    """Atualiza despesa. Retorna True se mudou alguma linha."""
    with connect() as conn:
        cur = conn.execute(
            "UPDATE expenses SET nome = ?, valor = ?, descricao = ? WHERE id = ?",
            (nome, valor, descricao, expense_id),
        )
        return cur.rowcount > 0


def zerar(expense_id: int) -> bool:
    """Seta valor=0. Retorna True se mudou alguma linha."""
    with connect() as conn:
        cur = conn.execute("UPDATE expenses SET valor = 0 WHERE id = ?", (expense_id,))
        return cur.rowcount > 0


def sum_all(excluding_id: Optional[int] = None) -> float:
    """Soma de todos os valores, opcionalmente excluindo um id (para edicao)."""
    with connect() as conn:
        if excluding_id is None:
            row = conn.execute(
                "SELECT COALESCE(SUM(valor), 0) AS s FROM expenses"
            ).fetchone()
        else:
            row = conn.execute(
                "SELECT COALESCE(SUM(valor), 0) AS s FROM expenses WHERE id != ?",
                (excluding_id,),
            ).fetchone()
        return float(row["s"])


def total_pages() -> int:
    """Numero total de paginas (>=1)."""
    n = count()
    if n == 0:
        return 1
    return (n + PAGE_SIZE - 1) // PAGE_SIZE
