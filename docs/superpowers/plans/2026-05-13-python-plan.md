# Python Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Python backend of estudoHtmx (lista de despesas com htmx) per the shared design spec, sharing `shared/expenses.db` with the other 4 stacks.

**Architecture:** FastAPI + Jinja2 server-rendered HTML, with htmx-driven partials returned as fragments. Stdlib `sqlite3` against the shared `shared/expenses.db` (modo WAL, schema idempotente). A custom sandboxed expression evaluator usa `ast` para `=expr` no campo valor; nenhum ORM nem libs de htmx no backend.

**Tech Stack:** Python 3.11+, FastAPI, Uvicorn, Jinja2, python-multipart, sqlite3 (stdlib), uv (package manager). Porta 5003. Frontend usa pico CSS + bootstrap icons via CDN.

---

### Task 1: Project setup (uv init + deps)

**Files:**
- Create: `python/pyproject.toml` (gerado por `uv init`)
- Create: `python/.python-version` (gerado por `uv init`)
- Create: `python/.gitignore`

- [ ] **Step 1: Criar a pasta `python/` e inicializar com uv**

Pre-requisito: ter `uv` instalado (https://docs.astral.sh/uv/). Caso nao tenha, fallback documentado no README usa `python -m venv` + `pip`.

Run from `C:\Dev\github\estudoHtmx`:
```powershell
mkdir python
cd python
uv init --no-readme --python 3.11
```

Isso gera `pyproject.toml`, `.python-version` e um `hello.py` placeholder. Apague o placeholder:

```powershell
Remove-Item hello.py -Force
```

- [ ] **Step 2: Adicionar dependencias runtime**

Run from `C:\Dev\github\estudoHtmx\python`:
```powershell
uv add fastapi uvicorn jinja2 python-multipart
```

Isso atualiza `pyproject.toml` e cria/atualiza `uv.lock`.

- [ ] **Step 3: Criar `.gitignore` da pasta python**

Create file `C:\Dev\github\estudoHtmx\python\.gitignore`:

```gitignore
__pycache__/
*.py[cod]
*$py.class
.venv/
.env
.pytest_cache/
.mypy_cache/
.ruff_cache/
*.egg-info/
```

- [ ] **Step 4: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/pyproject.toml python/.python-version python/uv.lock python/.gitignore
git commit -m "chore(python): setup do projeto com uv e dependencias fastapi"
```

---

### Task 2: SQLite connection helper + ensure schema

**Files:**
- Create: `python/db.py`

- [ ] **Step 1: Criar helper de conexao com schema idempotente**

Le `../shared/schema.sql` e executa via `executescript`. Cada request abre/fecha sua propria conexao (sqlite3 nao e thread-safe entre threads do FastAPI por padrao).

Create file `C:\Dev\github\estudoHtmx\python\db.py`:

```python
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
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/db.py
git commit -m "feat(python): conexao sqlite3 e bootstrap idempotente do schema"
```

---

### Task 3: Domain type — `Expense`

**Files:**
- Create: `python/models.py`

- [ ] **Step 1: Criar dataclass Expense**

Dataclass simples (sem Pydantic — fica leve, e os dados vem direto do sqlite3.Row).

Create file `C:\Dev\github\estudoHtmx\python\models.py`:

```python
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
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/models.py
git commit -m "feat(python): dataclass expense de dominio"
```

---

### Task 4: Data access (repository)

**Files:**
- Create: `python/repository.py`

- [ ] **Step 1: Implementar todas as operacoes CRUD + window function**

20 itens por pagina (constante PAGE_SIZE). running sum global via `SUM(valor) OVER (ORDER BY id)`.

Create file `C:\Dev\github\estudoHtmx\python\repository.py`:

```python
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
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/repository.py
git commit -m "feat(python): repository com paginacao window function e operacoes crud"
```

---

### Task 5: Expression parser via `ast` (sandboxed)

**Files:**
- Create: `python/parser.py`

- [ ] **Step 1: Implementar parser custom usando `ast`**

Aceita numero literal ou expressao com prefixo `=`. Numero literal: ponto OU virgula como decimal. Expressao: somente ponto como decimal (sintaxe Python valida). Walka o AST permitindo so um set restrito de nodes — qualquer outro node levanta erro.

Create file `C:\Dev\github\estudoHtmx\python\parser.py`:

```python
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
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/parser.py
git commit -m "feat(python): parser ast sandboxed pro campo valor"
```

---

### Task 6: Currency formatting (BRL pt-BR)

**Files:**
- Create: `python/format.py`

- [ ] **Step 1: Implementar formatter BRL sem depender de locale do sistema**

A funcao `brl` formata `R$ 1.234,56`, com negativos como `-R$ 50,00`. Tambem retorna uma classe CSS auxiliar pra cor (negativo = vermelho).

Create file `C:\Dev\github\estudoHtmx\python\format.py`:

```python
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
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/format.py
git commit -m "feat(python): formatador brl sem dependencia de locale"
```

---

### Task 7: Layout template (pico CSS + bootstrap icons)

**Files:**
- Create: `python/templates/layout.html`

- [ ] **Step 1: Criar template base com pico + bootstrap icons via CDN**

Container centralizado, mobile-friendly. Define dois IDs alvos: `#expense-form` (form) e `#expense-list-wrapper` (a tabela inteira incluindo paginacao). Sao os targets que o htmx vai trocar.

Create file `C:\Dev\github\estudoHtmx\python\templates\layout.html`:

```html
<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>estudoHtmx — python</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.min.css">
  <script src="https://unpkg.com/htmx.org@1.9.12"></script>
  <style>
    .negativo { color: var(--pico-color-red-500, #c0392b); }
    .acoes button { padding: 0.25rem 0.5rem; margin: 0 0.1rem; }
    .pagination { display: flex; gap: 0.25rem; justify-content: center; align-items: center; margin-top: 1rem; }
    .pagination button { padding: 0.25rem 0.6rem; }
    .pagination .pag-info { padding: 0 0.5rem; }
    small.error { color: var(--pico-color-red-500, #c0392b); display: block; margin-top: 0.25rem; }
    small.hint { color: var(--pico-muted-color, #6c757d); display: block; margin-top: 0.25rem; }
    td.truncate { max-width: 18rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    main.container { max-width: 960px; }
  </style>
</head>
<body>
  <main class="container">
    <hgroup>
      <h1>despesas</h1>
      <p>estudoHtmx — backend python (fastapi + jinja2)</p>
    </hgroup>

    <section id="form-wrapper">
      {% include "form.html" %}
    </section>

    <section id="expense-list-wrapper">
      {% include "list.html" %}
    </section>
  </main>
</body>
</html>
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/templates/layout.html
git commit -m "feat(python): template layout com pico css e bootstrap icons"
```

---

### Task 8: Form fragment

**Files:**
- Create: `python/templates/form.html`

- [ ] **Step 1: Criar fragmento do form com hidden id, hint do valor, total_acumulado readonly via htmx**

O contexto Jinja deve fornecer:
- `expense` (dict ou None) — quando editando
- `errors` (dict campo->mensagem) — vazio na maioria dos casos
- `total_acumulado_str` (string ja formatada) — soma excluindo o id em edicao
- `is_oob` (bool) — se True, marca `hx-swap-oob="true"` (usado no POST)

Create file `C:\Dev\github\estudoHtmx\python\templates\form.html`:

```html
{# fragmento: <form id="expense-form"> #}
<form id="expense-form"
      hx-post="/expenses"
      hx-target="#expense-form"
      hx-swap="outerHTML"
      {% if is_oob %}hx-swap-oob="true"{% endif %}>

  <input type="hidden" name="id" value="{{ (expense.id if expense else '') }}">

  <label>
    nome
    <input type="text" name="nome"
           value="{{ (expense.nome if expense else '') }}"
           required minlength="1"
           aria-invalid="{{ 'true' if errors.get('nome') else 'false' }}">
    {% if errors.get('nome') %}<small class="error">{{ errors['nome'] }}</small>{% endif %}
  </label>

  <label>
    valor
    <input type="text" name="valor"
           value="{{ (raw_valor if raw_valor is defined else (expense.valor if expense else '')) }}"
           required
           aria-invalid="{{ 'true' if errors.get('valor') else 'false' }}">
    <small class="hint">ex: 100, -50, 1234,56 ou =(50/2)+10</small>
    {% if errors.get('valor') %}<small class="error">{{ errors['valor'] }}</small>{% endif %}
  </label>

  <label>
    total acumulado
    <input id="total_acumulado" name="total_acumulado" type="text" readonly
           value="{{ total_acumulado_str }}"
           hx-get="/expenses/total-context?excluding_id={{ expense.id if expense else '' }}"
           hx-trigger="focusout from:#expense-form input delay:300ms, focusout from:#expense-form textarea delay:300ms"
           hx-target="this"
           hx-swap="outerHTML">
  </label>

  <label>
    descricao
    <textarea name="descricao" rows="2">{{ (expense.descricao if expense else '') }}</textarea>
  </label>

  <div class="grid">
    <button type="submit">salvar</button>
    <button type="button" class="secondary"
            hx-get="/expenses/form"
            hx-target="#expense-form"
            hx-swap="outerHTML">cancelar</button>
  </div>
</form>
```

- [ ] **Step 2: Criar template do fragmento total_acumulado isolado**

Create file `C:\Dev\github\estudoHtmx\python\templates\total_context.html`:

```html
{# retornado por GET /expenses/total-context — substitui o proprio <input> #}
<input id="total_acumulado" name="total_acumulado" type="text" readonly
       value="{{ total_acumulado_str }}"
       hx-get="/expenses/total-context?excluding_id={{ excluding_id or '' }}"
       hx-trigger="focusout from:#expense-form input delay:300ms, focusout from:#expense-form textarea delay:300ms"
       hx-target="this"
       hx-swap="outerHTML">
```

- [ ] **Step 3: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/templates/form.html python/templates/total_context.html
git commit -m "feat(python): fragmento do form com hint e total acumulado via htmx"
```

---

### Task 9: List fragment (tabela + paginacao)

**Files:**
- Create: `python/templates/list.html`

- [ ] **Step 1: Criar fragmento da lista com tabela e controles de paginacao**

Contexto Jinja: `expenses` (lista), `page` (int), `total_pages` (int), `is_oob` (bool).

Create file `C:\Dev\github\estudoHtmx\python\templates\list.html`:

```html
{# fragmento: a tabela inteira da pagina atual. id principal e expense-list-wrapper. #}
<div id="expense-list-wrapper"
     {% if is_oob %}hx-swap-oob="true"{% endif %}>
  <table>
    <thead>
      <tr>
        <th>#</th>
        <th>nome</th>
        <th>valor</th>
        <th>total acumulado</th>
        <th>descricao</th>
        <th>acoes</th>
      </tr>
    </thead>
    <tbody id="expense-list">
      {% if expenses|length == 0 %}
        <tr><td colspan="6"><em>nenhuma despesa cadastrada ainda.</em></td></tr>
      {% else %}
        {% for e in expenses %}
        <tr>
          <td>{{ e.id }}</td>
          <td>{{ e.nome }}</td>
          <td class="{{ 'negativo' if e.valor < 0 else '' }}">{{ e.valor|brl }}</td>
          <td class="{{ 'negativo' if e.total_acumulado < 0 else '' }}">{{ e.total_acumulado|brl }}</td>
          <td class="truncate" title="{{ e.descricao }}">{{ e.descricao }}</td>
          <td class="acoes">
            <button type="button"
                    title="editar"
                    hx-get="/expenses/form?id={{ e.id }}"
                    hx-target="#expense-form"
                    hx-swap="outerHTML">
              <i class="bi bi-pencil"></i>
            </button>
            <button type="button"
                    class="secondary"
                    title="zerar"
                    hx-post="/expenses/{{ e.id }}/zerar"
                    hx-confirm="Zerar este item?"
                    hx-target="#expense-list-wrapper"
                    hx-swap="outerHTML">
              <i class="bi bi-arrow-counterclockwise"></i>
            </button>
          </td>
        </tr>
        {% endfor %}
      {% endif %}
    </tbody>
  </table>

  <nav class="pagination" aria-label="paginacao">
    <button type="button"
            {% if page <= 1 %}disabled{% endif %}
            hx-get="/expenses?page=1"
            hx-target="#expense-list-wrapper"
            hx-swap="outerHTML"
            title="primeira">&laquo;</button>
    <button type="button"
            {% if page <= 1 %}disabled{% endif %}
            hx-get="/expenses?page={{ page - 1 }}"
            hx-target="#expense-list-wrapper"
            hx-swap="outerHTML"
            title="anterior">&lt;</button>
    <span class="pag-info">pagina {{ page }} de {{ total_pages }}</span>
    <button type="button"
            {% if page >= total_pages %}disabled{% endif %}
            hx-get="/expenses?page={{ page + 1 }}"
            hx-target="#expense-list-wrapper"
            hx-swap="outerHTML"
            title="proxima">&gt;</button>
    <button type="button"
            {% if page >= total_pages %}disabled{% endif %}
            hx-get="/expenses?page={{ total_pages }}"
            hx-target="#expense-list-wrapper"
            hx-swap="outerHTML"
            title="ultima">&raquo;</button>
  </nav>
</div>
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/templates/list.html
git commit -m "feat(python): fragmento da lista com paginacao e acoes"
```

---

### Task 10: Route GET /

**Files:**
- Create: `python/main.py`

- [ ] **Step 1: Criar app FastAPI com startup, templates e rota raiz**

Este passo cria o esqueleto e a primeira rota. Os proximos passos adicionam as demais rotas no MESMO arquivo. Mantenha o conteudo na ordem indicada para facilitar leitura.

Create file `C:\Dev\github\estudoHtmx\python\main.py`:

```python
"""estudoHtmx — backend python (fastapi + jinja2). Porta 5003."""
from __future__ import annotations

from pathlib import Path
from typing import Optional

from fastapi import FastAPI, Form, Request, Response
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates

import repository
from db import ensure_schema
from format import brl
from parser import ParseError, parse_valor

BASE_DIR = Path(__file__).resolve().parent
templates = Jinja2Templates(directory=str(BASE_DIR / "templates"))
templates.env.filters["brl"] = brl

app = FastAPI(title="estudoHtmx python")


@app.on_event("startup")
def _startup() -> None:
    ensure_schema()


def _is_htmx(request: Request) -> bool:
    return request.headers.get("hx-request", "").lower() == "true"


def _render(name: str, request: Request, ctx: dict) -> HTMLResponse:
    ctx_full = {"request": request, **ctx}
    return templates.TemplateResponse(name, ctx_full)


@app.get("/", response_class=HTMLResponse)
def index(request: Request):
    """Pagina completa: form vazio + lista pag 1."""
    expenses = repository.list_page(1)
    total_pages = repository.total_pages()
    total_str = brl(repository.sum_all(None))
    ctx = {
        "expense": None,
        "errors": {},
        "total_acumulado_str": total_str,
        "is_oob": False,
        "expenses": expenses,
        "page": 1,
        "total_pages": total_pages,
    }
    return _render("layout.html", request, ctx)
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/main.py
git commit -m "feat(python): app fastapi com startup e rota raiz"
```

---

### Task 11: Route GET /expenses?page=N

**Files:**
- Modify: `python/main.py`

- [ ] **Step 1: Adicionar handler que retorna o fragmento da lista**

Edit file `C:\Dev\github\estudoHtmx\python\main.py` — adicione ao final:

```python
@app.get("/expenses", response_class=HTMLResponse)
def list_expenses(request: Request, page: int = 1):
    """Fragmento da lista paginada."""
    if page < 1:
        page = 1
    total_pages = repository.total_pages()
    if page > total_pages:
        page = total_pages
    expenses = repository.list_page(page)
    ctx = {
        "expenses": expenses,
        "page": page,
        "total_pages": total_pages,
        "is_oob": False,
    }
    return _render("list.html", request, ctx)
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/main.py
git commit -m "feat(python): rota get expenses com paginacao"
```

---

### Task 12: Route GET /expenses/form?id=N

**Files:**
- Modify: `python/main.py`

- [ ] **Step 1: Adicionar handler que retorna o fragmento do form (vazio ou preenchido)**

Edit file `C:\Dev\github\estudoHtmx\python\main.py` — adicione ao final:

```python
@app.get("/expenses/form", response_class=HTMLResponse)
def form_fragment(request: Request, id: Optional[int] = None):
    """Form vazio (modo novo) ou preenchido pra edicao."""
    expense = repository.get_by_id(id) if id else None
    excluding = expense.id if expense else None
    total_str = brl(repository.sum_all(excluding))
    ctx = {
        "expense": expense,
        "errors": {},
        "total_acumulado_str": total_str,
        "is_oob": False,
    }
    return _render("form.html", request, ctx)
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/main.py
git commit -m "feat(python): rota get expenses form para novo ou edicao"
```

---

### Task 13: Route GET /expenses/total-context?excluding_id=N

**Files:**
- Modify: `python/main.py`

- [ ] **Step 1: Adicionar handler que retorna so o input do total_acumulado**

Edit file `C:\Dev\github\estudoHtmx\python\main.py` — adicione ao final:

```python
@app.get("/expenses/total-context", response_class=HTMLResponse)
def total_context(request: Request, excluding_id: Optional[int] = None):
    """Fragmento <input> com soma de todos os valores, exceto excluding_id."""
    # excluding_id pode chegar como '' por causa do template — Optional[int] lida com isso
    total = repository.sum_all(excluding_id)
    ctx = {
        "total_acumulado_str": brl(total),
        "excluding_id": excluding_id,
    }
    return _render("total_context.html", request, ctx)
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/main.py
git commit -m "feat(python): rota get total-context com soma excluindo id"
```

---

### Task 14: Route POST /expenses

**Files:**
- Modify: `python/main.py`

- [ ] **Step 1: Adicionar handler de submit com validacao, parse e resposta com OOB swap**

A resposta concatena dois fragmentos: o `<form>` re-renderizado (alvo principal) e o `<div id="expense-list-wrapper">` via `hx-swap-oob="true"` (alvo secundario). Quando ha erro de validacao, retorna apenas o form com erros — sem OOB.

Edit file `C:\Dev\github\estudoHtmx\python\main.py` — adicione ao final:

```python
@app.post("/expenses", response_class=HTMLResponse)
async def submit_expense(
    request: Request,
    id: str = Form(""),
    nome: str = Form(""),
    valor: str = Form(""),
    descricao: str = Form(""),
):
    """Submit do form. Valida, persiste, retorna form + tbody (oob)."""
    errors: dict[str, str] = {}

    # parse id (vazio = novo)
    expense_id: Optional[int] = None
    if id.strip():
        try:
            expense_id = int(id)
        except ValueError:
            expense_id = None  # ignora id invalido, trata como novo

    # validacao nome
    nome_clean = nome.strip()
    if nome_clean == "":
        errors["nome"] = "nome obrigatorio"

    # validacao valor
    valor_parsed: Optional[float] = None
    try:
        valor_parsed = parse_valor(valor)
    except ParseError as e:
        errors["valor"] = str(e)

    descricao_clean = descricao or ""

    if errors:
        # re-renderiza o form com erros — preserva inputs e raw_valor
        excluding = expense_id
        total_str = brl(repository.sum_all(excluding))
        # monta um "pseudo-expense" pra repreencher (evita perder dados em caso de erro)
        partial_expense = None
        if expense_id is not None:
            existing = repository.get_by_id(expense_id)
            if existing:
                partial_expense = existing
        # injeta nome/descricao digitados no contexto via dict-like
        # truque simples: usar um SimpleNamespace
        from types import SimpleNamespace
        expense_ctx = SimpleNamespace(
            id=expense_id if expense_id is not None else "",
            nome=nome_clean,
            valor=valor,  # raw — pra preservar o que o usuario digitou
            descricao=descricao_clean,
            created_at="",
        )
        ctx = {
            "expense": expense_ctx,
            "errors": errors,
            "total_acumulado_str": total_str,
            "is_oob": False,
            "raw_valor": valor,
        }
        return _render("form.html", request, ctx)

    # persiste
    assert valor_parsed is not None
    if expense_id is not None and repository.get_by_id(expense_id) is not None:
        repository.update(expense_id, nome_clean, valor_parsed, descricao_clean)
    else:
        repository.insert(nome_clean, valor_parsed, descricao_clean)

    # response: form zerado + lista oob da pagina 1
    page = 1
    total_pages = repository.total_pages()
    expenses = repository.list_page(page)
    total_str = brl(repository.sum_all(None))

    form_html = templates.get_template("form.html").render({
        "request": request,
        "expense": None,
        "errors": {},
        "total_acumulado_str": total_str,
        "is_oob": False,
    })
    list_html = templates.get_template("list.html").render({
        "request": request,
        "expenses": expenses,
        "page": page,
        "total_pages": total_pages,
        "is_oob": True,
    })

    combined = form_html + "\n" + list_html
    return HTMLResponse(
        content=combined,
        headers={"HX-Trigger": "itemSaved"},
    )
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/main.py
git commit -m "feat(python): rota post expenses com validacao parse e oob swap"
```

---

### Task 15: Route POST /expenses/{id}/zerar

**Files:**
- Modify: `python/main.py`

- [ ] **Step 1: Adicionar handler que zera valor e retorna lista atualizada**

A response retorna o fragmento da lista (alvo principal do botao zerar) e, OOB, o form re-renderizado caso o item zerado seja o que esta em edicao. Como nao temos estado server-side de "qual item esta em edicao", sempre retornamos o form em modo novo via OOB — simplificacao aceitavel (o usuario que estava editando ve o form limpar, sinal claro de que a acao tomou efeito).

Edit file `C:\Dev\github\estudoHtmx\python\main.py` — adicione ao final:

```python
@app.post("/expenses/{expense_id}/zerar", response_class=HTMLResponse)
def zerar_expense(request: Request, expense_id: int):
    """Seta valor=0. Retorna lista atualizada + form limpo (oob)."""
    repository.zerar(expense_id)

    page = 1
    total_pages = repository.total_pages()
    expenses = repository.list_page(page)
    total_str = brl(repository.sum_all(None))

    list_html = templates.get_template("list.html").render({
        "request": request,
        "expenses": expenses,
        "page": page,
        "total_pages": total_pages,
        "is_oob": False,
    })
    form_html = templates.get_template("form.html").render({
        "request": request,
        "expense": None,
        "errors": {},
        "total_acumulado_str": total_str,
        "is_oob": True,
    })

    combined = list_html + "\n" + form_html
    return HTMLResponse(
        content=combined,
        headers={"HX-Trigger": "itemZerado"},
    )
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/main.py
git commit -m "feat(python): rota post zerar com lista atualizada e form oob"
```

---

### Task 16: README do `python/`

**Files:**
- Create: `python/README.md`

- [ ] **Step 1: Documentar install (uv + fallback pip), run, porta 5003, e acceptance test**

Create file `C:\Dev\github\estudoHtmx\python\README.md`:

````markdown
# estudoHtmx — backend python

backend python (fastapi + jinja2 + sqlite3) do estudoHtmx. porta **5003**.

compartilha `../shared/expenses.db` com os outros stacks.

## requisitos

- python 3.11+
- [uv](https://docs.astral.sh/uv/) (recomendado) — gerenciador de pacotes rapido

## instalar e rodar (com uv)

a partir desta pasta (`python/`):

```powershell
uv sync
uv run uvicorn main:app --port 5003 --reload
```

acesse: http://localhost:5003

## fallback (sem uv: venv + pip)

```powershell
python -m venv .venv
.venv\Scripts\Activate.ps1
pip install fastapi uvicorn jinja2 python-multipart
uvicorn main:app --port 5003 --reload
```

## estrutura

```
python/
├── main.py            # app fastapi + rotas
├── db.py              # conexao sqlite3 + ensure schema
├── repository.py      # crud + paginacao window function
├── parser.py          # parser ast sandboxed pro campo valor
├── format.py          # formatador brl
├── models.py          # dataclass expense
├── templates/
│   ├── layout.html         # pagina completa
│   ├── form.html           # fragmento do form
│   ├── list.html           # fragmento da lista (tabela + paginacao)
│   └── total_context.html  # fragmento do input total_acumulado
├── pyproject.toml
└── uv.lock
```

## decisoes

- **sem ORM**: stdlib `sqlite3` direto. schema vem de `../shared/schema.sql`.
- **sem lib htmx**: header `hx-request` checado manualmente. swap oob feito
  concatenando HTML no response.
- **parser custom via `ast`**: walka a arvore validando so `BinOp(+,-,*,/)`,
  `UnaryOp(+,-)` e `Constant` (int/float). Qualquer outro node e rejeitado
  antes do `eval`. globals/locals vazios — sem builtins.

## acceptance test manual

- [ ] subir o servidor, abrir http://localhost:5003, ver pagina completa
- [ ] criar uma despesa com valor literal (ex: `100`)
- [ ] criar uma despesa com expressao (ex: `=(-1*(50/2)+10)` -> `-15,00`)
- [ ] editar a despesa criada (clicar editar, alterar, salvar)
- [ ] validar erro: salvar com nome vazio mostra erro inline
- [ ] validar erro: salvar com `valor=abc` mostra erro inline
- [ ] zerar uma despesa (item continua na lista com valor 0, soma 0)
- [ ] criar 25 despesas, navegar pra pagina 2, running sum continua corretamente
- [ ] total_acumulado do form se atualiza ao tabular pra fora do form
- [ ] dados persistem ao trocar de stack (mesmo banco)
````

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:
```powershell
git add python/README.md
git commit -m "docs(python): readme com install run e acceptance test"
```

---

### Task 17: Acceptance test checklist (manual)

**Files:**
- (sem arquivos novos — checklist a executar manualmente apos rodar o app)

- [ ] **Step 1: Subir o servidor**

Run from `C:\Dev\github\estudoHtmx\python`:
```powershell
uv run uvicorn main:app --port 5003 --reload
```

Em outro terminal, verifica que o banco foi criado:
```powershell
Test-Path C:\Dev\github\estudoHtmx\shared\expenses.db
```

- [ ] **Step 2: Smoke test no browser (http://localhost:5003)**

Abrir http://localhost:5003 e confirmar:
- pagina renderiza com form vazio em cima + tabela vazia embaixo
- pico CSS aplicado (container centralizado, font system)
- icones bootstrap carregam (vai aparecer quando tiver linhas)

- [ ] **Step 3: Criar despesa com valor literal**

No form: nome=`almoco`, valor=`50,75`, descricao=`segunda`. Clicar **salvar**.

Confirmar:
- form volta vazio
- linha aparece na tabela: `R$ 50,75`, total acumulado `R$ 50,75`

- [ ] **Step 4: Criar despesa com expressao**

No form: nome=`compra`, valor=`=(-1*(50/2)+10)`, descricao=`teste expr`. Clicar **salvar**.

Confirmar:
- valor salvo como `-R$ 15,00` (em vermelho)
- total acumulado da linha = `R$ 35,75`

- [ ] **Step 5: Editar despesa**

Clicar `bi-pencil` na primeira linha. Form preenche. Alterar valor pra `=50,75+0`.

Note: isso deve falhar (virgula em expressao). Clicar salvar -> esperado erro inline `sintaxe invalida` no campo valor.

Trocar pra `=50.75` -> salvar -> sucesso.

- [ ] **Step 6: Validar erro: nome vazio**

Cancelar edicao. No form: nome=` ` (so espacos), valor=`10`. Salvar.

Confirmar: erro inline `nome obrigatorio` abaixo do nome.

- [ ] **Step 7: Validar erro: valor invalido**

No form: nome=`teste`, valor=`abc`. Salvar.

Confirmar: erro inline `valor invalido` abaixo do valor.

- [ ] **Step 8: Zerar despesa**

Clicar `bi-arrow-counterclockwise` em uma linha. Confirmar dialog `Zerar este item?`. Clicar ok.

Confirmar:
- linha continua na tabela com `R$ 0,00`
- total acumulado recalculado

- [ ] **Step 9: Paginacao + running sum global**

Criar 25 despesas (qualquer valor, ex: `10`). Navegar pra pagina 2.

Confirmar:
- 5 linhas na pag 2
- total acumulado da linha 21 = soma das 20 anteriores + valor proprio (running sum continua)
- botoes `<<` e `<` desabilitados na pag 1; `>` e `>>` desabilitados na pag 2

- [ ] **Step 10: total_acumulado focusout**

Limpar form (cancelar). Digitar nome `x`, valor `100`, sair do campo (Tab). Apos ~300ms, total_acumulado atualiza com soma de tudo.

Editar uma despesa existente. total_acumulado mostra soma **excluindo** ela.

- [ ] **Step 11: Persistencia cross-stack**

Parar o servidor python. Confirmar que `shared/expenses.db` contem os dados:
```powershell
sqlite3 C:\Dev\github\estudoHtmx\shared\expenses.db "SELECT COUNT(*) FROM expenses;"
```

(quando os outros stacks existirem, subir outro stack e ver os mesmos dados.)

- [ ] **Step 12: Commit final do checklist (opcional)**

Se durante o teste manual algum bug for encontrado e corrigido, commit os fixes. Caso contrario, nenhum commit nesta task.
