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
