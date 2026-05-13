# estudoHtmx

estudo comparativo de 5 backends servindo o mesmo aplicativo de despesas via htmx,
compartilhando um unico banco sqlite. objetivo: comparar produtividade, simplicidade
de codigo e curva de aprendizado dos 5 stacks pra app simples server-driven com htmx.

## status

todas as 5 implementacoes estao **prontas e mergeadas em `main`**. cada uma
passou smoke test (HTTP 200 com HTML valido: `<form>`, tabela, hint, pico CSS,
htmx) e compartilha o mesmo `shared/expenses.db`. cada stack tem seu proprio
`README.md` na subpasta com detalhes de install/run.

branches `feat/<stack>` preservadas no remoto pra diff/historico isolado.

## stacks comparados

| stack | tecnologias | porta |
|---|---|---|
| .net | razor pages + ef core + htmx + htmx.taghelpers | 5001 |
| go | net/http + html/template + database/sql + modernc.org/sqlite | 5002 |
| python | fastapi + jinja2 + sqlite3 | 5003 |
| node | hono + jsx + better-sqlite3 (via node + tsx) | 5004 |
| rust | axum + askama + rusqlite | 5005 |

## como rodar

todas as stacks leem/escrevem em `shared/expenses.db` (criado no primeiro start
a partir de `shared/schema.sql`).

**rodar uma por vez** (sqlite com WAL aguenta multipla leitura, mas pra evitar
lock contention durante o estudo a recomendacao e ligar apenas um servidor de
cada vez). use Ctrl+C pra parar antes de subir o proximo.

a partir da raiz do repositorio. **4 das 5 stacks tem comandos identicos
em qualquer plataforma** (so o node difere por causa de bug do bun no windows):

```sh
# .net     ->  http://localhost:5001
cd dotnet
dotnet run

# go       ->  http://localhost:5002
cd go
go run .

# python (com uv)  ->  http://localhost:5003
cd python
uv run uvicorn main:app --port 5003

# rust     ->  http://localhost:5005
cd rust
cargo run
```

### node — comando depende da plataforma

linux / mac (bun funciona normalmente):

```sh
cd node
bun --hot run server.tsx       # http://localhost:5004
```

windows (use Node + tsx — bun segfalha aqui):

```powershell
cd node
npx tsx server.tsx             # http://localhost:5004
```

> motivo: `bun 1.1.x` crasha ao carregar `better-sqlite3` no windows
> (incompatibilidade conhecida do `dlopen` do bun com o N-API binding nativo).
> em linux/mac nao acontece.

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

apos rodar e exercitar manualmente as 5 stacks, escrever um `COMPARISON.md`
na raiz com:

- tabela quantitativa (loc, arquivos, deps, tempos)
- impressoes qualitativas por stack (o que foi facil, o que foi chato)
- ranking subjetivo de **produtividade** e **simplicidade**
- escolha final recomendada pra "projeto simples + htmx"

## design

ver [doc de design](docs/superpowers/specs/2026-05-13-htmx-stacks-design.md).
cada stack tem seu proprio plan de implementacao em
`docs/superpowers/plans/2026-05-13-<stack>-plan.md`.
