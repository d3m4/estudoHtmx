# estudoHtmx — design doc

**data:** 2026-05-13
**autor:** felipe
**status:** aguardando aprovacao

---

## 1. visao geral

construir o mesmo aplicativo simples de **lista de despesas** em 5 backends
diferentes, todos servindo HTML server-rendered + interatividade via htmx, e
todos compartilhando o **mesmo arquivo sqlite**. o app e identico em UX e
comportamento — o que varia e o stack por tras.

## 2. objetivos do estudo

- comparar **produtividade real** dos 5 stacks pro padrao server-driven + htmx
- avaliar **simplicidade de codigo** (linhas, arquivos, conceitos exigidos)
- avaliar **curva de aprendizado** vs ergonomia da rotina diaria
- validar a tese de **validacao server-side via htmx** (single source of truth
  pras regras, sem duplicar validacao em JS) em todas as stacks

## 3. dominio

uma unica entidade: **expense** (despesa).

```
expense
  id          inteiro auto-increment, pk
  nome        texto, obrigatorio
  valor       numerico, obrigatorio (aceita negativo e zero)
  descricao   texto, opcional (default '')
  created_at  timestamp, auto
```

### regras de negocio

- **zerar** em vez de deletar: nao ha exclusao nesta versao. o usuario "zera" um
  registro, o que seta `valor = 0`. o registro continua existindo (visivel na
  lista, com valor 0, contribuindo 0 pra running sum).
- **editar**: form com `id` no hidden field. submit com id existente = update,
  sem id = insert.
- **valor**: aceita expressao aritmetica quando string comeca com `=`. ex:
  `=(-1*(50/2)+10)` -> `-15`. avaliada no backend antes de persistir.

## 4. ui / ux

### 4.1 layout

uma unica pagina centralizada (container do pico css), mobile-friendly:

```
+----------------------+
| FORM (criar/editar)  |
+----------------------+
| LISTA (paginada)     |
| 20 itens por pagina  |
+----------------------+
```

ordem: form **em cima**, lista **embaixo**.

### 4.2 formulario

campos, na ordem:

| campo | tipo | obrigatorio | observacoes |
|---|---|---|---|
| id | hidden | nao | preenchido quando editando |
| nome | text | sim | min 1 char |
| valor | text | sim | aceita numero direto ou expressao `=...`. **hint visual obrigatorio**: um `<small>` abaixo do input com texto `ex: 100, -50, 1234,56 ou =(50/2)+10` (cor neutra/muted do pico) |
| total_acumulado | text readonly | n/a | carregado via ajax (ver 4.4) |
| descricao | textarea | nao | livre |

botoes: **salvar** (submit) e **cancelar** (limpa form e volta pro modo "novo").

### 4.3 tabela de despesas

colunas, na ordem:

| coluna | conteudo |
|---|---|
| # | id |
| nome | nome |
| valor | formatado br: `R$ 1.234,56` (negativos em vermelho) |
| total acumulado | running sum global por id ASC, formatado br |
| descricao | descricao (truncada com tooltip se longa) |
| acoes | **editar** (icone `bi-pencil`) e **zerar** (icone `bi-arrow-counterclockwise`) |

icones via cdn de **bootstrap icons**.

#### paginacao

- 20 itens por pagina
- ordenacao fixa: `id ASC`
- controles: `[<<] [<] pagina X de Y [>] [>>]`
- a running sum (total acumulado) e **global**: linha 21 (primeira da pag 2)
  comeca de onde a pag 1 terminou. implementacao no backend via window
  function do sqlite:
  ```sql
  SELECT id, nome, valor, descricao,
         SUM(valor) OVER (ORDER BY id) AS total_acumulado
    FROM expenses
   ORDER BY id
   LIMIT 20 OFFSET ?
  ```
  isso resolve o running sum global em um unico query, independente de pagina.

### 4.4 total acumulado no form

- campo readonly
- carregado via **ajax htmx no `focusout` do form** com debounce 300ms:
  ```html
  <input id="total_acumulado" readonly
         hx-get="/expenses/total-context?excluding_id={id ou vazio}"
         hx-trigger="focusout from:input delay:300ms,
                     focusout from:textarea delay:300ms"
         hx-target="this"
         hx-swap="outerHTML">
  ```
- semantica: soma de todos os `expense.valor` **exceto** o registro sendo
  editado (quando id presente). se criando novo (sem id), e a soma de todos
  os registros.
- formatado como BRL.

### 4.5 locale e formatacao

- locale: **pt-BR**
- moeda: `R$ 1.234,56` (ponto milhar, virgula decimal, simbolo na frente)
- decimais negativos: `-R$ 50,00`, em cor `var(--pico-color-red-500)` ou similar
- entrada do campo `valor`:
  - **numero literal** (sem `=`): aceita ponto OU virgula como decimal. ex:
    `100,50` e `100.50` -> 100.5
  - **expressao** (com `=`): convencao **ponto como decimal** (compatibilidade
    com sintaxe matematica padrao). virgula reservada pra futura extensao
    (separador de args de funcao). ex: `=1.5+2.5` ok; `=1,5+2,5` invalido

## 5. arquitetura

### 5.1 estrutura de pastas

```
estudoHtmx/
├── README.md
├── .gitignore
├── docs/
│   └── superpowers/
│       └── specs/
│           └── 2026-05-13-htmx-stacks-design.md
├── shared/
│   ├── schema.sql           # canonical, idempotente
│   └── expenses.db          # gerado em runtime, gitignored
├── dotnet/                  # razor pages + ef core + htmx
├── go/                      # net/http + html/template + database/sql
├── python/                  # fastapi + jinja2
├── node/                    # hono + jsx
└── rust/                    # axum + askama
```

### 5.2 banco compartilhado

- arquivo unico: `shared/expenses.db` (sqlite, modo WAL)
- schema em `shared/schema.sql` com `CREATE TABLE IF NOT EXISTS` — idempotente
- cada stack, no startup, **executa o schema.sql** uma vez (no-op se ja
  existe). zero ORM migrations, zero drift
- decisao: **rodar uma stack por vez** durante o estudo, pra evitar
  contencao de lock e simplificar comparacao

#### schema.sql (rascunho)

```sql
CREATE TABLE IF NOT EXISTS expenses (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    nome        TEXT    NOT NULL CHECK(length(trim(nome)) > 0),
    valor       REAL    NOT NULL DEFAULT 0,
    descricao   TEXT    NOT NULL DEFAULT '',
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_expenses_id ON expenses(id);
```

nota: `valor` como REAL (decimal flutuante) por simplicidade. armazenar como
INTEGER de centavos seria mais correto financeiramente, mas a UX/parser fica
mais limpo com REAL. para um estudo, ok.

### 5.3 portas

| stack | porta | motivo |
|---|---|---|
| .net | 5001 | padrao kestrel http dev |
| go | 5002 | sequencial |
| python | 5003 | sequencial |
| node | 5004 | sequencial |
| rust | 5005 | sequencial |

## 6. endpoints (identicos nos 5 stacks)

| metodo | rota | resposta |
|---|---|---|
| GET | `/` | pagina completa (form vazio + lista pag 1) |
| GET | `/expenses?page=N` | fragmento htmx: `<table>` da pagina N (incluindo paginacao) |
| GET | `/expenses/form?id=N` | fragmento htmx: `<form>` preenchido pra edicao (id opcional) |
| GET | `/expenses/total-context?excluding_id=N` | fragmento htmx: `<input>` do total_acumulado |
| POST | `/expenses` | submit do form. retorna `<form>` re-renderizado + `<tbody>` da lista via `hx-swap-oob` |
| POST | `/expenses/{id}/zerar` | seta valor=0, retorna `<tbody>` recalculado via oob + `<form>` re-renderizado se for o item em edicao. botao no front usa `hx-confirm="Zerar este item?"` |

### 6.1 fluxo de submit (POST /expenses)

1. backend recebe form data (id?, nome, valor, descricao)
2. valida:
   - nome: obrigatorio, trim != ""
   - valor: parse (se comeca com `=`, avalia expressao). erro de sintaxe vira erro de validacao
   - descricao: livre (opcional)
3. se invalido: retorna `<form>` com mensagens de erro inline (mantem inputs preenchidos). HTTP 200 (htmx swap normal).
4. se valido:
   - id presente -> UPDATE
   - id ausente -> INSERT
   - retorna `<form>` zerado (modo "novo") + `<tbody>` recalculado da pagina atual via `hx-swap-oob="true"`
5. flash msg opcional: `HX-Trigger: itemSaved` pro front mostrar toast

### 6.2 trigger oob

o response do POST contem:

```html
<!-- target principal: o form -->
<form id="expense-form" hx-swap-oob="true">
  ... form zerado ou com erros ...
</form>

<!-- swap fora de banda: a tabela -->
<tbody id="expense-list" hx-swap-oob="true">
  ... linhas recalculadas ...
</tbody>
```

front configurado com `hx-target="#expense-form"` no submit; o `<tbody>` se
atualiza sozinho pelo oob.

## 7. parser de expressao no campo valor

### 7.1 escopo

- aritmetica basica: `+`, `-`, `*`, `/`, `(`, `)`
- numeros: inteiros, decimais (ponto OU virgula), negativos
- prefixo `=` opcional: se ausente, trata como numero direto
- **sem** funcoes, sem refs cross-row, sem variaveis

### 7.2 exemplos

| input | resultado |
|---|---|
| `100` | 100.0 |
| `100,50` | 100.5 |
| `100.50` | 100.5 |
| `-50` | -50.0 |
| `=50+10` | 60.0 |
| `=(-1*(50/2)+10)` | -15.0 |
| `=1500/12` | 125.0 |
| `=10/0` | erro: divisao por zero |
| `=abs(5)` | erro: funcao nao suportada |
| `abc` | erro: nao e numero nem expressao |

### 7.3 implementacao por stack

cada stack escolhe a abordagem mais idiomatica:

| stack | abordagem |
|---|---|
| .net | `System.Data.DataTable().Compute()` (built-in, suporta todos os operadores) |
| go | `github.com/expr-lang/expr` (avaliador seguro, sandboxed) |
| python | parser custom usando `ast` (parse + walk validando so os AST nodes permitidos: `BinOp`, `UnaryOp`, `Num`, `Constant`) |
| node | `expr-eval` ou parser custom |
| rust | crate `evalexpr` (sandboxed por padrao) |

**seguranca**: todos os parsers devem ser sandbox (sem `eval` de codigo do
host). e input do usuario.

## 8. validacao

a validacao acontece **so no submit do form** (modelo on-commit, conforme
discutido — sem validacao por-campo no blur).

regras:

- `nome`: trim, len >= 1
- `valor`: parse (numero ou expressao) -> deve render numero finito
- `descricao`: nenhuma validacao (default '')

erros sao **devolvidos como HTML** no proprio fragmento do form
(`<small class="error">...</small>` abaixo do input invalido).

## 9. acoes do usuario (fluxo completo)

| acao | request | response |
|---|---|---|
| abrir app | `GET /` | pagina inteira |
| clicar **editar** numa linha | `GET /expenses/form?id=X` (`hx-target="#expense-form"`) | form preenchido |
| clicar **cancelar** | `GET /expenses/form` (sem id) | form vazio |
| editar valor e sair do form (focusout) | `GET /expenses/total-context?excluding_id=X` | input do total_acumulado atualizado |
| clicar **salvar** | `POST /expenses` (form data) | form atualizado + tbody oob |
| clicar **zerar** numa linha | `POST /expenses/X/zerar` (com confirm dialog) | tr atualizada via oob |
| clicar **paginacao** | `GET /expenses?page=N` (`hx-target="#expense-list"`) | tabela inteira da pagina N |

## 10. criterios de aceitacao

por stack:

- [ ] subir o servidor, abrir no browser, ver pagina completa
- [ ] criar uma despesa com valor literal (ex: `100`)
- [ ] criar uma despesa com expressao (ex: `=(-1*(50/2)+10)`)
- [ ] editar a despesa criada (clicar editar, alterar, salvar)
- [ ] validar erro: salvar com nome vazio mostra erro inline
- [ ] validar erro: salvar com `valor=abc` mostra erro inline
- [ ] zerar uma despesa (item continua na lista com valor 0, soma 0)
- [ ] criar 25 despesas, navegar pra pagina 2, running sum continua corretamente
- [ ] total_acumulado do form se atualiza ao tabular pra fora do form
- [ ] dados persistem ao trocar de stack (mesmo banco)

cross-stack:

- [ ] todos os 5 stacks veem os mesmos dados do `shared/expenses.db`
- [ ] running sum, paginacao e formatacao BRL identicos visualmente

## 11. fora de escopo (nao fazer nesta versao)

- autenticacao / multi-usuario
- delete real (so zerar)
- expressoes com funcoes ou refs cross-row
- export / import
- testes automatizados (avaliar adicionar numa segunda rodada)
- docker / containerizacao
- ci / cd
- tema dark / customizacao visual alem do pico default
- i18n alem de pt-BR

## 12. riscos e mitigacoes

| risco | mitigacao |
|---|---|
| lock contention no sqlite com 2+ stacks rodando | rodar uma por vez; modo WAL habilitado no schema |
| parser de expressao com bug de seguranca (eval malicioso) | usar libs sandbox conhecidas, nunca `eval()` direto |
| drift de schema entre stacks | schema canonical em sql, todas usam o mesmo arquivo |
| running sum performance com muitos registros | nao e foco; sqlite + window function `SUM() OVER (ORDER BY id)` resolve trivialmente |
| valor REAL e impreciso pra dinheiro | aceito pro estudo; documentado |

## 13. estimativa de esforco

por stack, aproximadamente:

- setup do projeto + dependencias: 15-30 min
- conexao sqlite + ensure schema: 15 min
- rotas + handlers: 1-2 h
- templates: 1-2 h
- parser de expressao + validacao: 30 min
- ajustes finais + teste manual: 30 min

**total por stack: ~3-5h.** 5 stacks: ~15-25h de trabalho. recomendado dividir
em sessoes — uma stack por sessao.

ordem sugerida de implementacao:

1. **.net** (mais familiaridade do autor, baseline)
2. **go** (contraste compilado-minimalista)
3. **python** (contraste dinamico-batteries-included)
4. **node** (jsx + bun)
5. **rust** (mais complexo, deixar pro final)
