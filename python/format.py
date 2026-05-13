"""Formatacao BRL pt-BR (sem depender de locale do sistema)."""
from __future__ import annotations


def brl(value: float) -> str:
    """Formata float como moeda BRL: 'R$ 1.234,56' ou '-R$ 50,00'."""
    if value is None:
        value = 0.0
    negative = value < 0
    n = abs(float(value))
    # arredonda pra 2 casas
    s = f"{n:,.2f}"  # ex: '1,234.56'
    # troca separadores: ',' -> placeholder, '.' -> ',', placeholder -> '.'
    s = s.replace(",", "_").replace(".", ",").replace("_", ".")
    sign = "-" if negative else ""
    return f"{sign}R$ {s}"


def is_negative(value: float) -> bool:
    """Helper para template (classe css)."""
    return float(value or 0) < 0
