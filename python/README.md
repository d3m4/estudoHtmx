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
