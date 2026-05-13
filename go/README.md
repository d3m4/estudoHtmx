# go — estudoHtmx

backend Go do projeto estudoHtmx. usa stdlib `net/http` (ServeMux com
patterns Go 1.22+), `html/template`, `database/sql` sobre
`modernc.org/sqlite` (pure-Go, sem CGo) e `github.com/expr-lang/expr` pro
parser de expressao no campo valor.

## requisitos

- Go **1.22 ou superior** (necessario para ServeMux com path patterns
  tipo `POST /expenses/{id}/zerar`)
- nenhum compilador C externo (driver sqlite e pure-Go)

## instalar e rodar

de dentro de `go/`:

```powershell
go mod tidy        # baixa as dependencias
go run .           # sobe o servidor na porta 5002
```

ou compilar e rodar o binario:

```powershell
go build -o estudohtmx-go.exe .
.\estudohtmx-go.exe
```

acessar: <http://localhost:5002>

## como funciona

- no startup, abre/cria `../shared/expenses.db` e executa
  `../shared/schema.sql` idempotentemente (CREATE TABLE IF NOT EXISTS)
- rotas (todas servindo HTML server-rendered + htmx):
  - `GET /` — pagina completa (form + lista paginada)
  - `GET /expenses?page=N` — fragmento da tabela com 20 itens/pagina
  - `GET /expenses/form?id=N` — fragmento do form (vazio ou pre-preenchido pra editar)
  - `GET /expenses/total-context?excluding_id=N` — input do total acumulado via AJAX
  - `POST /expenses` — submit do form; valida; insert ou update; retorna form + tbody via `hx-swap-oob`
  - `POST /expenses/{id}/zerar` — seta valor=0; retorna tbody via OOB
- graceful shutdown via SIGINT/SIGTERM com timeout de 5s

## acceptance test manual

1. subir o servidor: `go run .`
2. abrir <http://localhost:5002>
3. criar uma despesa com valor literal `100` -> aparece como `R$ 100,00`
4. criar uma despesa com expressao `=(-1*(50/2)+10)` -> aparece como `-R$ 15,00` em vermelho
5. clicar **editar** numa linha -> form preenche
6. alterar nome, **salvar** -> tabela atualiza
7. salvar com nome vazio -> erro inline `nome obrigatorio` no campo
8. salvar com valor `abc` -> erro inline `valor invalido` no campo
9. clicar **zerar** -> confirm `Zerar este item?` -> linha permanece com `R$ 0,00`
10. criar 25 despesas, ir pra pagina 2 -> running sum continua de onde a pagina 1 parou
11. focar em um input, tabular pra fora -> apos 300ms o total acumulado se atualiza
12. parar e subir outra stack apontando pro mesmo `shared/expenses.db` -> mesmos dados visiveis
