# estudoHtmx

estudo comparativo de 5 backends servindo o mesmo aplicativo de despesas via htmx,
compartilhando um unico banco sqlite. objetivo: comparar produtividade, simplicidade
de codigo e curva de aprendizado dos 5 stacks pra app simples server-driven com htmx.

## stacks comparados

| stack | tecnologias | porta |
|---|---|---|
| .net | razor pages + ef core + htmx + htmx.taghelpers | 5001 |
| go | net/http + html/template + database/sql + modernc.org/sqlite | 5002 |
| python | fastapi + jinja2 + sqlite3 | 5003 |
| node | hono + jsx + better-sqlite3 | 5004 |
| rust | axum + askama + rusqlite | 5005 |

## como rodar

cada subpasta tem seu proprio readme. todas leem/escrevem em `shared/expenses.db`
(criado no primeiro start a partir de `shared/schema.sql`).

**rodar uma por vez** (sqlite com wal aguenta, mas pra evitar lock contention durante
o estudo a recomendacao e ligar apenas um servidor de cada vez).

## design

ver [doc de design](docs/superpowers/specs/2026-05-13-htmx-stacks-design.md).
