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

cada subpasta tem seu proprio readme com o comando especifico. todas leem/escrevem
em `shared/expenses.db` (criado no primeiro start a partir de `shared/schema.sql`).

**rodar uma por vez** (sqlite com wal aguenta, mas pra evitar lock contention durante
o estudo a recomendacao e ligar apenas um servidor de cada vez).

acessar no browser na porta correspondente da tabela acima.

## como comparar

cada stack expoe **a mesma UI e o mesmo comportamento**, lendo e escrevendo o
mesmo `shared/expenses.db`. esse e o ponto central do estudo: variavel unica =
o stack. fluxo recomendado:

1. subir uma stack (ex: `cd dotnet && dotnet run`)
2. exercitar as funcionalidades pelo browser:
   - criar despesa com valor literal (`100`, `1234,56`)
   - criar despesa com expressao (`=(50/2)+10`)
   - editar uma despesa existente (clicar editar)
   - zerar uma despesa (item permanece com valor 0)
   - paginar (criar 25+ despesas, ir pra pag 2, validar running sum continua)
   - validar erros (nome vazio, valor invalido)
3. parar a stack, subir a proxima (ex: `cd ../go && go run .`)
4. **as mesmas despesas aparecem** (mesmo db)
5. repetir pros 5 stacks

### aspectos a observar e medir

| dimensao | como medir |
|---|---|
| linhas de codigo | `tokei <pasta>` ou `cloc <pasta>` |
| numero de arquivos | `Get-ChildItem -Recurse \| Measure-Object` |
| dependencias externas | contar entradas no `*.csproj` / `go.mod` / `requirements.txt` / `package.json` / `Cargo.toml` |
| tempo de cold start | medir do `run` ate primeira request servir |
| tempo de hot reload | editar template, observar se atualiza sem restart |
| boilerplate de htmx | linhas de codigo dedicadas a integrar htmx (headers, partials, etc) |
| clareza do handler de submit | ler o codigo de `POST /expenses` e avaliar densidade conceitual |
| ergonomia do parser de expressao | foi trivial (lib pronta) ou exigiu parser custom? |
| facilidade de retornar fragmentos html | como cada stack constroi/retorna partials |
| tipagem ate o template | template e checked em compile time (.net, rust) ou runtime (go, python, node)? |

### entregavel da comparacao

ao fim das 5 implementacoes, escrever um `COMPARISON.md` na raiz com:

- tabela quantitativa (loc, arquivos, deps, tempos)
- impressoes qualitativas por stack (o que foi facil, o que foi chato)
- ranking subjetivo de **produtividade** e **simplicidade**
- escolha final recomendada pra "projeto simples + htmx"

## design

ver [doc de design](docs/superpowers/specs/2026-05-13-htmx-stacks-design.md).
cada stack tem seu proprio plan de implementacao em
`docs/superpowers/plans/2026-05-13-<stack>-plan.md`.
