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
    # starlette 0.35+ exige request como primeiro arg posicional
    return templates.TemplateResponse(request, name, ctx)


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
