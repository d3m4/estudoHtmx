"""Parser do campo valor: numero literal ou expressao com prefixo '='.

Implementacao: usa `ast.parse(expr, mode='eval')` e valida o AST permitindo
apenas BinOp(+,-,*,/), UnaryOp(+,-), Constant(int/float) e Expression. Qualquer
outro node levanta ValueError. Isso e seguro porque NUNCA chamamos `eval` sobre
input bruto: so apos walk-and-validate completo.
"""
from __future__ import annotations

import ast
import math
from typing import Final

_ALLOWED_BINOPS: Final = (ast.Add, ast.Sub, ast.Mult, ast.Div)
_ALLOWED_UNARYOPS: Final = (ast.UAdd, ast.USub)


class ParseError(ValueError):
    """Erro de parsing do campo valor."""


def parse_valor(raw: str) -> float:
    """Converte string do campo valor em float.

    - vazio -> ParseError
    - comeca com '=' -> avalia como expressao matematica sandboxed
    - senao -> tenta parsear como numero literal (aceita ',' ou '.' como decimal)

    Levanta ParseError em qualquer caso invalido.
    """
    if raw is None:
        raise ParseError("valor obrigatorio")
    s = raw.strip()
    if s == "":
        raise ParseError("valor obrigatorio")

    if s.startswith("="):
        expr = s[1:].strip()
        if expr == "":
            raise ParseError("expressao vazia apos '='")
        return _eval_expr(expr)

    # numero literal: aceita virgula OU ponto como decimal
    normalized = s.replace(",", ".")
    try:
        v = float(normalized)
    except ValueError:
        raise ParseError("valor invalido")
    if not math.isfinite(v):
        raise ParseError("valor invalido")
    return v


def _eval_expr(expr: str) -> float:
    """Avalia expressao matematica sandboxed via ast."""
    try:
        tree = ast.parse(expr, mode="eval")
    except SyntaxError as e:
        raise ParseError(f"sintaxe invalida: {e.msg}") from e

    _validate(tree)

    try:
        code = compile(tree, "<valor>", "eval")
        # globals/locals vazios — sem builtins, sem nada
        result = eval(code, {"__builtins__": {}}, {})  # noqa: S307
    except ZeroDivisionError as e:
        raise ParseError("divisao por zero") from e
    except Exception as e:  # qualquer outro erro de runtime
        raise ParseError(f"erro ao avaliar: {e}") from e

    if not isinstance(result, (int, float)):
        raise ParseError("resultado nao e numerico")
    f = float(result)
    if not math.isfinite(f):
        raise ParseError("resultado nao finito")
    return f


def _validate(node: ast.AST) -> None:
    """Walk recursivo permitindo apenas nodes seguros."""
    if isinstance(node, ast.Expression):
        _validate(node.body)
        return
    if isinstance(node, ast.BinOp):
        if not isinstance(node.op, _ALLOWED_BINOPS):
            raise ParseError(f"operador nao suportado: {type(node.op).__name__}")
        _validate(node.left)
        _validate(node.right)
        return
    if isinstance(node, ast.UnaryOp):
        if not isinstance(node.op, _ALLOWED_UNARYOPS):
            raise ParseError(f"operador unario nao suportado: {type(node.op).__name__}")
        _validate(node.operand)
        return
    if isinstance(node, ast.Constant):
        if not isinstance(node.value, (int, float)) or isinstance(node.value, bool):
            raise ParseError(f"constante nao suportada: {node.value!r}")
        return
    # qualquer outro node (Name, Call, Attribute, Subscript, etc) -> rejeita
    raise ParseError(f"expressao contem elemento nao permitido: {type(node).__name__}")
