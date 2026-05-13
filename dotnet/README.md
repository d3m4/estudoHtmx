# estudoHtmx — backend .NET

Implementacao do estudoHtmx em ASP.NET Core 9 Razor Pages + HTMX,
compartilhando `shared/expenses.db` com os outros 4 stacks.

## requisitos

- .NET SDK 9.x instalado (`dotnet --version` >= 9.0.0)

## instalar dependencias

A partir da raiz do repo (`estudoHtmx`):

```powershell
dotnet restore dotnet/EstudoHtmx.csproj
```

## rodar

```powershell
dotnet run --project dotnet/EstudoHtmx.csproj
```

A app sobe em `http://localhost:5001`.

No startup, o app aplica `../shared/schema.sql` no banco
`../shared/expenses.db` (cria o arquivo se nao existir). A operacao
e idempotente (usa `CREATE TABLE IF NOT EXISTS`).

## estrutura

```
dotnet/
  EstudoHtmx.csproj
  Program.cs
  appsettings.json
  Data/
    SqliteOptions.cs
    SqliteConnectionFactory.cs
    SchemaInitializer.cs
    ExpenseRepository.cs
  Domain/
    Expense.cs
  Services/
    ValorParser.cs
    CurrencyFormatter.cs
  Pages/
    _ViewImports.cshtml
    _ViewStart.cshtml
    Index.cshtml
    Index.cshtml.cs
    Shared/
      _Layout.cshtml
      _Form.cshtml
      _FormModel.cs
      _List.cshtml
      _ListModel.cs
      _Tbody.cshtml
      _TotalContext.cshtml
      _TotalContextModel.cs
      _FormWithOob.cshtml
      _FormWithOobModel.cs
    Expenses/
      Index.cshtml          # GET /expenses?page=N
      Form.cshtml           # GET /expenses/form?id=N
      TotalContext.cshtml   # GET /expenses/total-context?excluding_id=N
      Submit.cshtml         # POST /expenses
      Zerar.cshtml          # POST /expenses/{id}/zerar
```

## rotas

| metodo | rota | resposta |
|---|---|---|
| GET  | `/`                                          | pagina completa |
| GET  | `/expenses?page=N`                           | fragmento da lista |
| GET  | `/expenses/form?id=N`                        | fragmento do form |
| GET  | `/expenses/total-context?excluding_id=N`     | input do total |
| POST | `/expenses`                                  | form + tbody via OOB |
| POST | `/expenses/{id}/zerar`                       | form + tbody via OOB |
