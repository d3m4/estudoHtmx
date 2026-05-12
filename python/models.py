"""Tipos de dominio."""
from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class Expense:
    id: int
    nome: str
    valor: float
    descricao: str
    created_at: str
    total_acumulado: float = 0.0  # preenchido nas queries com window function

    @classmethod
    def from_row(cls, row) -> "Expense":
        """Constroi a partir de sqlite3.Row. row pode ou nao ter total_acumulado."""
        keys = row.keys() if hasattr(row, "keys") else []
        return cls(
            id=row["id"],
            nome=row["nome"],
            valor=float(row["valor"]),
            descricao=row["descricao"],
            created_at=row["created_at"],
            total_acumulado=float(row["total_acumulado"]) if "total_acumulado" in keys else 0.0,
        )
