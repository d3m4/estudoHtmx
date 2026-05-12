# .NET Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the .NET backend of estudoHtmx (lista de despesas com htmx) per the shared design spec, sharing `shared/expenses.db` with the other 4 stacks.

**Architecture:** ASP.NET Core 9 Razor Pages app exposing one page (`/`) plus six HTMX fragment endpoints. Data access is hand-rolled `Microsoft.Data.Sqlite` (no EF Core migrations) because the schema is owned by `shared/schema.sql` and applied idempotently on startup. HTMX integration uses the `Htmx` and `Htmx.TagHelpers` NuGet packages by Khalid Abuhakmeh for `Request.IsHtmx()`/`Response.Htmx(...)` and `hx-*` tag helpers.

**Tech Stack:** .NET 9 SDK, ASP.NET Core Razor Pages, `Microsoft.Data.Sqlite` 9.x, `Htmx` 1.8.x, `Htmx.TagHelpers` 1.8.x, pico CSS via CDN, Bootstrap Icons via CDN.

---

### Task 1: Project setup

**Files:**
- Create: `dotnet/EstudoHtmx.csproj`
- Create: `dotnet/.gitignore`
- Create: `dotnet/appsettings.json`
- Create: `dotnet/appsettings.Development.json`
- Create: `dotnet/Properties/launchSettings.json`

- [ ] **Step 1: Scaffold the Razor Pages project**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet new webapp -n EstudoHtmx -o dotnet --framework net9.0
```

- [ ] **Step 2: Remove scaffolded example pages we won't use**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
Remove-Item -Recurse -Force dotnet/Pages/Privacy.cshtml, dotnet/Pages/Privacy.cshtml.cs, dotnet/Pages/Error.cshtml, dotnet/Pages/Error.cshtml.cs, dotnet/wwwroot/css, dotnet/wwwroot/js, dotnet/wwwroot/lib, dotnet/wwwroot/favicon.ico -ErrorAction SilentlyContinue
```

- [ ] **Step 3: Install NuGet dependencies**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet add dotnet/EstudoHtmx.csproj package Microsoft.Data.Sqlite --version 9.0.0
dotnet add dotnet/EstudoHtmx.csproj package Htmx --version 1.8.1
dotnet add dotnet/EstudoHtmx.csproj package Htmx.TagHelpers --version 1.8.1
```

- [ ] **Step 4: Replace `dotnet/EstudoHtmx.csproj` with a clean explicit version**

Replace the file `dotnet/EstudoHtmx.csproj` with:

```xml
<Project Sdk="Microsoft.NET.Sdk.Web">

  <PropertyGroup>
    <TargetFramework>net9.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <RootNamespace>EstudoHtmx</RootNamespace>
    <AssemblyName>EstudoHtmx</AssemblyName>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Htmx" Version="1.8.1" />
    <PackageReference Include="Htmx.TagHelpers" Version="1.8.1" />
    <PackageReference Include="Microsoft.Data.Sqlite" Version="9.0.0" />
  </ItemGroup>

</Project>
```

- [ ] **Step 5: Fix launch profile to bind to port 5001**

Replace `dotnet/Properties/launchSettings.json` with:

```json
{
  "profiles": {
    "EstudoHtmx": {
      "commandName": "Project",
      "dotnetRunMessages": true,
      "launchBrowser": false,
      "applicationUrl": "http://localhost:5001",
      "environmentVariables": {
        "ASPNETCORE_ENVIRONMENT": "Development"
      }
    }
  }
}
```

- [ ] **Step 6: Trim appsettings**

Replace `dotnet/appsettings.json` with:

```json
{
  "Logging": {
    "LogLevel": {
      "Default": "Information",
      "Microsoft.AspNetCore": "Warning"
    }
  },
  "AllowedHosts": "*",
  "SqliteDbPath": "../shared/expenses.db",
  "SqliteSchemaPath": "../shared/schema.sql"
}
```

Replace `dotnet/appsettings.Development.json` with:

```json
{
  "Logging": {
    "LogLevel": {
      "Default": "Information",
      "Microsoft.AspNetCore": "Warning"
    }
  }
}
```

- [ ] **Step 7: Add a project-level `.gitignore`**

Create `dotnet/.gitignore`:

```gitignore
bin/
obj/
*.user
.vs/
```

- [ ] **Step 8: Verify build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 9: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/EstudoHtmx.csproj dotnet/.gitignore dotnet/appsettings.json dotnet/appsettings.Development.json dotnet/Properties/launchSettings.json dotnet/Program.cs dotnet/Pages
git commit -m "feat(dotnet): inicializa projeto razor pages com htmx e sqlite"
```

---

### Task 2: SQLite connection + ensure schema on startup

**Files:**
- Create: `dotnet/Data/SqliteOptions.cs`
- Create: `dotnet/Data/SqliteConnectionFactory.cs`
- Create: `dotnet/Data/SchemaInitializer.cs`
- Modify: `dotnet/Program.cs`

- [ ] **Step 1: Create `dotnet/Data/SqliteOptions.cs`**

```csharp
namespace EstudoHtmx.Data;

public sealed class SqliteOptions
{
    public string SqliteDbPath { get; set; } = "../shared/expenses.db";
    public string SqliteSchemaPath { get; set; } = "../shared/schema.sql";
}
```

- [ ] **Step 2: Create `dotnet/Data/SqliteConnectionFactory.cs`**

```csharp
using Microsoft.Data.Sqlite;
using Microsoft.Extensions.Options;

namespace EstudoHtmx.Data;

public sealed class SqliteConnectionFactory
{
    private readonly string _connectionString;

    public SqliteConnectionFactory(IOptions<SqliteOptions> options, IHostEnvironment env)
    {
        var raw = options.Value.SqliteDbPath;
        var absolute = Path.IsPathRooted(raw)
            ? raw
            : Path.GetFullPath(Path.Combine(env.ContentRootPath, raw));

        _connectionString = new SqliteConnectionStringBuilder
        {
            DataSource = absolute,
            Mode = SqliteOpenMode.ReadWriteCreate,
            Cache = SqliteCacheMode.Shared,
            ForeignKeys = true,
        }.ToString();
    }

    public SqliteConnection Open()
    {
        var conn = new SqliteConnection(_connectionString);
        conn.Open();
        using (var pragma = conn.CreateCommand())
        {
            pragma.CommandText = "PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;";
            pragma.ExecuteNonQuery();
        }
        return conn;
    }
}
```

- [ ] **Step 3: Create `dotnet/Data/SchemaInitializer.cs`**

```csharp
using Microsoft.Extensions.Options;

namespace EstudoHtmx.Data;

public sealed class SchemaInitializer
{
    private readonly SqliteConnectionFactory _factory;
    private readonly string _schemaPath;
    private readonly ILogger<SchemaInitializer> _logger;

    public SchemaInitializer(
        SqliteConnectionFactory factory,
        IOptions<SqliteOptions> options,
        IHostEnvironment env,
        ILogger<SchemaInitializer> logger)
    {
        _factory = factory;
        _logger = logger;
        var raw = options.Value.SqliteSchemaPath;
        _schemaPath = Path.IsPathRooted(raw)
            ? raw
            : Path.GetFullPath(Path.Combine(env.ContentRootPath, raw));
    }

    public void EnsureSchema()
    {
        if (!File.Exists(_schemaPath))
        {
            throw new FileNotFoundException(
                $"schema.sql nao encontrado em {_schemaPath}");
        }

        var sql = File.ReadAllText(_schemaPath);
        using var conn = _factory.Open();
        using var cmd = conn.CreateCommand();
        cmd.CommandText = sql;
        cmd.ExecuteNonQuery();
        _logger.LogInformation("schema aplicado a partir de {Path}", _schemaPath);
    }
}
```

- [ ] **Step 4: Replace `dotnet/Program.cs` with the wiring**

```csharp
using EstudoHtmx.Data;

var builder = WebApplication.CreateBuilder(args);

builder.Services.Configure<SqliteOptions>(builder.Configuration);
builder.Services.AddSingleton<SqliteConnectionFactory>();
builder.Services.AddSingleton<SchemaInitializer>();

builder.Services.AddRazorPages();
builder.Services.AddAntiforgery(opts =>
{
    opts.HeaderName = "X-CSRF-TOKEN";
});

var app = builder.Build();

// aplica schema (idempotente) antes de aceitar requests
using (var scope = app.Services.CreateScope())
{
    scope.ServiceProvider
        .GetRequiredService<SchemaInitializer>()
        .EnsureSchema();
}

app.UseStaticFiles();
app.UseRouting();
app.UseAntiforgery();

app.MapRazorPages();

app.Run();
```

- [ ] **Step 5: Build to confirm wiring compiles**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 6: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Data/SqliteOptions.cs dotnet/Data/SqliteConnectionFactory.cs dotnet/Data/SchemaInitializer.cs dotnet/Program.cs
git commit -m "feat(dotnet): conecta sqlite compartilhado e aplica schema no startup"
```

---

### Task 3: Domain type — `Expense` entity

**Files:**
- Create: `dotnet/Domain/Expense.cs`

- [ ] **Step 1: Create `dotnet/Domain/Expense.cs`**

```csharp
namespace EstudoHtmx.Domain;

public sealed record Expense(
    long Id,
    string Nome,
    double Valor,
    string Descricao,
    string CreatedAt
);

public sealed record ExpenseRow(
    long Id,
    string Nome,
    double Valor,
    string Descricao,
    double TotalAcumulado
);
```

- [ ] **Step 2: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 3: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Domain/Expense.cs
git commit -m "feat(dotnet): adiciona record expense e expenserow"
```

---

### Task 4: Data access — repository with pagination, running sum, CRUD-like ops

**Files:**
- Create: `dotnet/Data/ExpenseRepository.cs`
- Modify: `dotnet/Program.cs`

- [ ] **Step 1: Create `dotnet/Data/ExpenseRepository.cs`**

```csharp
using EstudoHtmx.Domain;
using Microsoft.Data.Sqlite;

namespace EstudoHtmx.Data;

public sealed class ExpenseRepository
{
    public const int PageSize = 20;

    private readonly SqliteConnectionFactory _factory;

    public ExpenseRepository(SqliteConnectionFactory factory)
    {
        _factory = factory;
    }

    public int CountAll()
    {
        using var conn = _factory.Open();
        using var cmd = conn.CreateCommand();
        cmd.CommandText = "SELECT COUNT(*) FROM expenses;";
        var raw = cmd.ExecuteScalar();
        return Convert.ToInt32(raw ?? 0L);
    }

    public int PageCount()
    {
        var total = CountAll();
        if (total <= 0) return 1;
        return (int)Math.Ceiling(total / (double)PageSize);
    }

    public IReadOnlyList<ExpenseRow> ListPage(int page)
    {
        if (page < 1) page = 1;
        var offset = (page - 1) * PageSize;

        using var conn = _factory.Open();
        using var cmd = conn.CreateCommand();
        cmd.CommandText = @"
            SELECT id, nome, valor, descricao, total_acumulado
              FROM (
                SELECT id, nome, valor, descricao,
                       SUM(valor) OVER (ORDER BY id) AS total_acumulado
                  FROM expenses
                 ORDER BY id
              )
             ORDER BY id
             LIMIT $limit OFFSET $offset;
        ";
        cmd.Parameters.AddWithValue("$limit", PageSize);
        cmd.Parameters.AddWithValue("$offset", offset);

        var rows = new List<ExpenseRow>(PageSize);
        using var reader = cmd.ExecuteReader();
        while (reader.Read())
        {
            rows.Add(new ExpenseRow(
                Id: reader.GetInt64(0),
                Nome: reader.GetString(1),
                Valor: reader.GetDouble(2),
                Descricao: reader.GetString(3),
                TotalAcumulado: reader.GetDouble(4)
            ));
        }
        return rows;
    }

    public Expense? GetById(long id)
    {
        using var conn = _factory.Open();
        using var cmd = conn.CreateCommand();
        cmd.CommandText = @"
            SELECT id, nome, valor, descricao, created_at
              FROM expenses
             WHERE id = $id;
        ";
        cmd.Parameters.AddWithValue("$id", id);
        using var reader = cmd.ExecuteReader();
        if (!reader.Read()) return null;
        return new Expense(
            Id: reader.GetInt64(0),
            Nome: reader.GetString(1),
            Valor: reader.GetDouble(2),
            Descricao: reader.GetString(3),
            CreatedAt: reader.GetString(4)
        );
    }

    public long Insert(string nome, double valor, string descricao)
    {
        using var conn = _factory.Open();
        using var cmd = conn.CreateCommand();
        cmd.CommandText = @"
            INSERT INTO expenses (nome, valor, descricao)
            VALUES ($nome, $valor, $descricao);
            SELECT last_insert_rowid();
        ";
        cmd.Parameters.AddWithValue("$nome", nome);
        cmd.Parameters.AddWithValue("$valor", valor);
        cmd.Parameters.AddWithValue("$descricao", descricao);
        var raw = cmd.ExecuteScalar();
        return Convert.ToInt64(raw ?? 0L);
    }

    public bool Update(long id, string nome, double valor, string descricao)
    {
        using var conn = _factory.Open();
        using var cmd = conn.CreateCommand();
        cmd.CommandText = @"
            UPDATE expenses
               SET nome = $nome, valor = $valor, descricao = $descricao
             WHERE id = $id;
        ";
        cmd.Parameters.AddWithValue("$id", id);
        cmd.Parameters.AddWithValue("$nome", nome);
        cmd.Parameters.AddWithValue("$valor", valor);
        cmd.Parameters.AddWithValue("$descricao", descricao);
        return cmd.ExecuteNonQuery() == 1;
    }

    public bool Zerar(long id)
    {
        using var conn = _factory.Open();
        using var cmd = conn.CreateCommand();
        cmd.CommandText = "UPDATE expenses SET valor = 0 WHERE id = $id;";
        cmd.Parameters.AddWithValue("$id", id);
        return cmd.ExecuteNonQuery() == 1;
    }

    public double SumAll(long? excludingId = null)
    {
        using var conn = _factory.Open();
        using var cmd = conn.CreateCommand();
        if (excludingId.HasValue)
        {
            cmd.CommandText = @"
                SELECT COALESCE(SUM(valor), 0)
                  FROM expenses
                 WHERE id <> $id;
            ";
            cmd.Parameters.AddWithValue("$id", excludingId.Value);
        }
        else
        {
            cmd.CommandText = "SELECT COALESCE(SUM(valor), 0) FROM expenses;";
        }
        var raw = cmd.ExecuteScalar();
        if (raw is null || raw is DBNull) return 0d;
        return Convert.ToDouble(raw);
    }

    public int PageOfId(long id)
    {
        using var conn = _factory.Open();
        using var cmd = conn.CreateCommand();
        cmd.CommandText = @"
            SELECT COUNT(*) FROM expenses WHERE id <= $id;
        ";
        cmd.Parameters.AddWithValue("$id", id);
        var raw = cmd.ExecuteScalar();
        var pos = Convert.ToInt32(raw ?? 0L);
        if (pos <= 0) return 1;
        return (int)Math.Ceiling(pos / (double)PageSize);
    }
}
```

- [ ] **Step 2: Register the repository in `dotnet/Program.cs`**

In `dotnet/Program.cs`, find the line:

```csharp
builder.Services.AddSingleton<SchemaInitializer>();
```

And add immediately below it:

```csharp
builder.Services.AddSingleton<ExpenseRepository>();
```

- [ ] **Step 3: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 4: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Data/ExpenseRepository.cs dotnet/Program.cs
git commit -m "feat(dotnet): adiciona repositorio com paginacao e running sum"
```

---

### Task 5: Expression parser for the valor field

**Files:**
- Create: `dotnet/Services/ValorParser.cs`

- [ ] **Step 1: Create `dotnet/Services/ValorParser.cs`**

```csharp
using System.Data;
using System.Globalization;
using System.Text.RegularExpressions;

namespace EstudoHtmx.Services;

public static class ValorParser
{
    public readonly record struct ParseResult(bool Ok, double Value, string? Error);

    public static ParseResult Parse(string? raw)
    {
        if (string.IsNullOrWhiteSpace(raw))
            return new ParseResult(false, 0d, "valor obrigatorio");

        var trimmed = raw.Trim();

        if (trimmed.StartsWith("="))
        {
            var expr = trimmed[1..].Trim();
            if (expr.Length == 0)
                return new ParseResult(false, 0d, "expressao vazia");

            if (!Regex.IsMatch(expr, @"^[0-9\+\-\*\/\(\)\.\s]+$"))
                return new ParseResult(false, 0d,
                    "expressao invalida: use apenas + - * / ( ) e numeros com ponto decimal");

            try
            {
                var dt = new DataTable();
                var result = dt.Compute(expr, null);
                if (result is null || result is DBNull)
                    return new ParseResult(false, 0d, "expressao invalida");

                var d = Convert.ToDouble(result, CultureInfo.InvariantCulture);
                if (double.IsNaN(d) || double.IsInfinity(d))
                    return new ParseResult(false, 0d, "expressao invalida (divisao por zero?)");

                return new ParseResult(true, d, null);
            }
            catch (DivideByZeroException)
            {
                return new ParseResult(false, 0d, "divisao por zero");
            }
            catch (SyntaxErrorException)
            {
                return new ParseResult(false, 0d, "expressao com sintaxe invalida");
            }
            catch (EvaluateException)
            {
                return new ParseResult(false, 0d, "expressao nao pode ser avaliada");
            }
            catch (OverflowException)
            {
                return new ParseResult(false, 0d, "expressao gerou overflow");
            }
            catch (Exception)
            {
                return new ParseResult(false, 0d, "expressao invalida");
            }
        }

        // numero literal: aceita ponto OU virgula como decimal
        var normalized = trimmed.Replace(",", ".");
        if (!double.TryParse(
                normalized,
                NumberStyles.Float | NumberStyles.AllowLeadingSign,
                CultureInfo.InvariantCulture,
                out var n))
        {
            return new ParseResult(false, 0d, "valor invalido: nao e numero nem expressao");
        }

        if (double.IsNaN(n) || double.IsInfinity(n))
            return new ParseResult(false, 0d, "valor invalido");

        return new ParseResult(true, n, null);
    }
}
```

- [ ] **Step 2: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 3: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Services/ValorParser.cs
git commit -m "feat(dotnet): adiciona parser de valor com expressao aritmetica"
```

---

### Task 6: Currency formatting (BRL pt-BR)

**Files:**
- Create: `dotnet/Services/CurrencyFormatter.cs`

- [ ] **Step 1: Create `dotnet/Services/CurrencyFormatter.cs`**

```csharp
using System.Globalization;

namespace EstudoHtmx.Services;

public static class CurrencyFormatter
{
    private static readonly CultureInfo PtBr = new("pt-BR");

    public static string Brl(double value)
    {
        // pt-BR ja gera "R$ 1.234,56" e "-R$ 50,00"
        return value.ToString("C", PtBr);
    }

    public static bool IsNegative(double value) => value < 0d;
}
```

- [ ] **Step 2: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 3: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Services/CurrencyFormatter.cs
git commit -m "feat(dotnet): adiciona formatador brl pt-br"
```

---

### Task 7: Layout template — pico CSS + Bootstrap Icons + HTMX

**Files:**
- Modify: `dotnet/Pages/_ViewImports.cshtml`
- Modify: `dotnet/Pages/_ViewStart.cshtml`
- Modify: `dotnet/Pages/Shared/_Layout.cshtml`

- [ ] **Step 1: Replace `dotnet/Pages/_ViewImports.cshtml`**

```cshtml
@using EstudoHtmx
@using EstudoHtmx.Domain
@using EstudoHtmx.Services
@namespace EstudoHtmx.Pages
@addTagHelper *, Microsoft.AspNetCore.Mvc.TagHelpers
@addTagHelper *, Htmx.TagHelpers
```

- [ ] **Step 2: Ensure `dotnet/Pages/_ViewStart.cshtml` points at the layout**

Replace `dotnet/Pages/_ViewStart.cshtml` with:

```cshtml
@{
    Layout = "_Layout";
}
```

- [ ] **Step 3: Replace `dotnet/Pages/Shared/_Layout.cshtml`**

```cshtml
@inject Microsoft.AspNetCore.Antiforgery.IAntiforgery Antiforgery
@{
    var tokens = Antiforgery.GetAndStoreTokens(Context);
}
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>@ViewData["Title"] - estudoHtmx (.NET)</title>
    <link rel="stylesheet"
          href="https://cdn.jsdelivr.net/npm/@@picocss/pico@@2/css/pico.min.css" />
    <link rel="stylesheet"
          href="https://cdn.jsdelivr.net/npm/bootstrap-icons@@1.11.3/font/bootstrap-icons.min.css" />
    <script src="https://unpkg.com/htmx.org@@1.9.12"></script>
    <style>
        .valor-neg { color: var(--pico-color-red-500); }
        .hint { color: var(--pico-muted-color); }
        .acoes-cell { white-space: nowrap; }
        .acoes-cell button { padding: 0.25rem 0.5rem; margin-right: 0.25rem; }
        td.descricao { max-width: 16rem; overflow: hidden;
                       text-overflow: ellipsis; white-space: nowrap; }
        .pagination { display: flex; gap: 0.25rem; align-items: center;
                      justify-content: center; margin-top: 1rem; }
        .pagination button { padding: 0.25rem 0.6rem; }
    </style>
</head>
<body>
    <main class="container">
        <header>
            <h2>estudoHtmx — .NET</h2>
        </header>
        <input type="hidden" id="__rf" value="@tokens.RequestToken" />
        <script>
            document.body.addEventListener('htmx:configRequest', function (evt) {
                var t = document.getElementById('__rf');
                if (t) { evt.detail.headers['X-CSRF-TOKEN'] = t.value; }
            });
        </script>
        @RenderBody()
    </main>
</body>
</html>
```

- [ ] **Step 4: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 5: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Pages/_ViewImports.cshtml dotnet/Pages/_ViewStart.cshtml dotnet/Pages/Shared/_Layout.cshtml
git commit -m "feat(dotnet): layout com pico, bootstrap icons e antiforgery via htmx header"
```

---

### Task 8: Form fragment partial

**Files:**
- Create: `dotnet/Pages/Shared/_FormModel.cs`
- Create: `dotnet/Pages/Shared/_Form.cshtml`
- Create: `dotnet/Pages/Shared/_TotalContext.cshtml`
- Create: `dotnet/Pages/Shared/_TotalContextModel.cs`

- [ ] **Step 1: Create `dotnet/Pages/Shared/_FormModel.cs`**

```csharp
namespace EstudoHtmx.Pages.Shared;

public sealed class FormModel
{
    public long? Id { get; set; }
    public string Nome { get; set; } = "";
    public string ValorRaw { get; set; } = "";
    public string Descricao { get; set; } = "";
    public double TotalContexto { get; set; } = 0d;
    public string? ErroNome { get; set; }
    public string? ErroValor { get; set; }
}
```

- [ ] **Step 2: Create `dotnet/Pages/Shared/_TotalContextModel.cs`**

```csharp
namespace EstudoHtmx.Pages.Shared;

public sealed class TotalContextModel
{
    public long? ExcludingId { get; set; }
    public double Total { get; set; }
}
```

- [ ] **Step 3: Create `dotnet/Pages/Shared/_TotalContext.cshtml`**

```cshtml
@model EstudoHtmx.Pages.Shared.TotalContextModel
@{
    var idAttr = Model.ExcludingId.HasValue ? Model.ExcludingId.Value.ToString() : "";
}
<input id="total_acumulado"
       name="total_acumulado"
       type="text"
       readonly
       value="@EstudoHtmx.Services.CurrencyFormatter.Brl(Model.Total)"
       data-excluding-id="@idAttr"
       hx-get="/expenses/total-context?excluding_id=@idAttr"
       hx-trigger="focusout from:input delay:300ms, focusout from:textarea delay:300ms"
       hx-target="this"
       hx-swap="outerHTML" />
```

- [ ] **Step 4: Create `dotnet/Pages/Shared/_Form.cshtml`**

```cshtml
@model EstudoHtmx.Pages.Shared.FormModel
@{
    var idValue = Model.Id.HasValue ? Model.Id.Value.ToString() : "";
    var isEdit = Model.Id.HasValue;
}
<form id="expense-form"
      hx-post="/expenses"
      hx-target="#expense-form"
      hx-swap="outerHTML">
    <input type="hidden" id="expense-id" name="id" value="@idValue" />

    <label for="nome">
        nome
        <input type="text"
               id="nome"
               name="nome"
               value="@Model.Nome"
               aria-invalid="@(Model.ErroNome != null ? "true" : null)"
               required />
        @if (Model.ErroNome != null)
        {
            <small class="error" style="color: var(--pico-color-red-500);">
                @Model.ErroNome
            </small>
        }
    </label>

    <label for="valor">
        valor
        <input type="text"
               id="valor"
               name="valor"
               value="@Model.ValorRaw"
               aria-invalid="@(Model.ErroValor != null ? "true" : null)"
               required />
        <small class="hint">ex: 100, -50, 1234,56 ou =(50/2)+10</small>
        @if (Model.ErroValor != null)
        {
            <small class="error" style="color: var(--pico-color-red-500);">
                @Model.ErroValor
            </small>
        }
    </label>

    <label for="total_acumulado">
        total acumulado
        @await Html.PartialAsync("_TotalContext",
            new EstudoHtmx.Pages.Shared.TotalContextModel
            {
                ExcludingId = Model.Id,
                Total = Model.TotalContexto
            })
    </label>

    <label for="descricao">
        descricao
        <textarea id="descricao" name="descricao" rows="2">@Model.Descricao</textarea>
    </label>

    <div role="group">
        <button type="submit">salvar</button>
        <button type="button"
                class="secondary"
                hx-get="/expenses/form"
                hx-target="#expense-form"
                hx-swap="outerHTML">cancelar</button>
    </div>
    @if (isEdit)
    {
        <p><small>editando id @idValue</small></p>
    }
</form>
```

- [ ] **Step 5: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 6: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Pages/Shared/_FormModel.cs dotnet/Pages/Shared/_TotalContextModel.cs dotnet/Pages/Shared/_TotalContext.cshtml dotnet/Pages/Shared/_Form.cshtml
git commit -m "feat(dotnet): adiciona partial do form e do total acumulado"
```

---

### Task 9: List fragment partial

**Files:**
- Create: `dotnet/Pages/Shared/_ListModel.cs`
- Create: `dotnet/Pages/Shared/_List.cshtml`
- Create: `dotnet/Pages/Shared/_Tbody.cshtml`

- [ ] **Step 1: Create `dotnet/Pages/Shared/_ListModel.cs`**

```csharp
using EstudoHtmx.Domain;

namespace EstudoHtmx.Pages.Shared;

public sealed class ListModel
{
    public IReadOnlyList<ExpenseRow> Rows { get; set; } = Array.Empty<ExpenseRow>();
    public int Page { get; set; } = 1;
    public int TotalPages { get; set; } = 1;
}
```

- [ ] **Step 2: Create `dotnet/Pages/Shared/_Tbody.cshtml`**

```cshtml
@model EstudoHtmx.Pages.Shared.ListModel
<tbody id="expense-list">
    @if (Model.Rows.Count == 0)
    {
        <tr>
            <td colspan="6"><em>nenhuma despesa cadastrada</em></td>
        </tr>
    }
    else
    {
        foreach (var r in Model.Rows)
        {
            var valorClass = r.Valor < 0 ? "valor-neg" : "";
            var totalClass = r.TotalAcumulado < 0 ? "valor-neg" : "";
            <tr id="row-@r.Id">
                <td>@r.Id</td>
                <td>@r.Nome</td>
                <td class="@valorClass">@EstudoHtmx.Services.CurrencyFormatter.Brl(r.Valor)</td>
                <td class="@totalClass">@EstudoHtmx.Services.CurrencyFormatter.Brl(r.TotalAcumulado)</td>
                <td class="descricao" title="@r.Descricao">@r.Descricao</td>
                <td class="acoes-cell">
                    <button type="button"
                            class="secondary"
                            title="editar"
                            hx-get="/expenses/form?id=@r.Id"
                            hx-target="#expense-form"
                            hx-swap="outerHTML">
                        <i class="bi bi-pencil"></i>
                    </button>
                    <button type="button"
                            class="contrast"
                            title="zerar"
                            hx-post="/expenses/@r.Id/zerar"
                            hx-target="#expense-list"
                            hx-swap="outerHTML"
                            hx-confirm="Zerar este item?">
                        <i class="bi bi-arrow-counterclockwise"></i>
                    </button>
                </td>
            </tr>
        }
    }
</tbody>
```

- [ ] **Step 3: Create `dotnet/Pages/Shared/_List.cshtml`**

```cshtml
@model EstudoHtmx.Pages.Shared.ListModel
@{
    var prev = Math.Max(1, Model.Page - 1);
    var next = Math.Min(Model.TotalPages, Model.Page + 1);
    var atFirst = Model.Page <= 1;
    var atLast = Model.Page >= Model.TotalPages;
}
<div id="expense-list-wrapper">
    <figure>
        <table role="grid">
            <thead>
                <tr>
                    <th scope="col">#</th>
                    <th scope="col">nome</th>
                    <th scope="col">valor</th>
                    <th scope="col">total acumulado</th>
                    <th scope="col">descricao</th>
                    <th scope="col">acoes</th>
                </tr>
            </thead>
            @await Html.PartialAsync("_Tbody", Model)
        </table>
    </figure>

    <nav class="pagination" id="expense-pagination">
        <button type="button"
                class="secondary"
                @(atFirst ? "disabled" : "")
                hx-get="/expenses?page=1"
                hx-target="#expense-list-wrapper"
                hx-swap="outerHTML">&lt;&lt;</button>
        <button type="button"
                class="secondary"
                @(atFirst ? "disabled" : "")
                hx-get="/expenses?page=@prev"
                hx-target="#expense-list-wrapper"
                hx-swap="outerHTML">&lt;</button>
        <span>pagina @Model.Page de @Model.TotalPages</span>
        <button type="button"
                class="secondary"
                @(atLast ? "disabled" : "")
                hx-get="/expenses?page=@next"
                hx-target="#expense-list-wrapper"
                hx-swap="outerHTML">&gt;</button>
        <button type="button"
                class="secondary"
                @(atLast ? "disabled" : "")
                hx-get="/expenses?page=@Model.TotalPages"
                hx-target="#expense-list-wrapper"
                hx-swap="outerHTML">&gt;&gt;</button>
    </nav>
</div>
```

- [ ] **Step 4: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 5: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Pages/Shared/_ListModel.cs dotnet/Pages/Shared/_List.cshtml dotnet/Pages/Shared/_Tbody.cshtml
git commit -m "feat(dotnet): adiciona partials da lista com paginacao"
```

---

### Task 10: Route GET / — full page

**Files:**
- Modify: `dotnet/Pages/Index.cshtml`
- Modify: `dotnet/Pages/Index.cshtml.cs`

- [ ] **Step 1: Replace `dotnet/Pages/Index.cshtml.cs`**

```csharp
using EstudoHtmx.Data;
using EstudoHtmx.Pages.Shared;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace EstudoHtmx.Pages;

public class IndexModel : PageModel
{
    private readonly ExpenseRepository _repo;

    public IndexModel(ExpenseRepository repo)
    {
        _repo = repo;
    }

    public FormModel FormVm { get; private set; } = new();
    public ListModel ListVm { get; private set; } = new();

    public void OnGet()
    {
        var totalPages = _repo.PageCount();
        var rows = _repo.ListPage(1);
        var sumAll = _repo.SumAll(excludingId: null);

        FormVm = new FormModel
        {
            Id = null,
            Nome = "",
            ValorRaw = "",
            Descricao = "",
            TotalContexto = sumAll
        };
        ListVm = new ListModel
        {
            Rows = rows,
            Page = 1,
            TotalPages = totalPages
        };
    }
}
```

- [ ] **Step 2: Replace `dotnet/Pages/Index.cshtml`**

```cshtml
@page
@model EstudoHtmx.Pages.IndexModel
@{
    ViewData["Title"] = "despesas";
}

<section>
    @await Html.PartialAsync("_Form", Model.FormVm)
</section>

<section>
    @await Html.PartialAsync("_List", Model.ListVm)
</section>
```

- [ ] **Step 3: Build and smoke-run**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 4: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Pages/Index.cshtml dotnet/Pages/Index.cshtml.cs
git commit -m "feat(dotnet): rota get / renderiza pagina completa"
```

---

### Task 11: Route GET /expenses?page=N — list fragment

**Files:**
- Create: `dotnet/Pages/Expenses/Index.cshtml`
- Create: `dotnet/Pages/Expenses/Index.cshtml.cs`

- [ ] **Step 1: Create `dotnet/Pages/Expenses/Index.cshtml.cs`**

```csharp
using EstudoHtmx.Data;
using EstudoHtmx.Pages.Shared;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace EstudoHtmx.Pages.Expenses;

public class IndexModel : PageModel
{
    private readonly ExpenseRepository _repo;

    public IndexModel(ExpenseRepository repo)
    {
        _repo = repo;
    }

    public ListModel ListVm { get; private set; } = new();

    public IActionResult OnGet([FromQuery] int page = 1)
    {
        var totalPages = _repo.PageCount();
        if (page < 1) page = 1;
        if (page > totalPages) page = totalPages;

        ListVm = new ListModel
        {
            Rows = _repo.ListPage(page),
            Page = page,
            TotalPages = totalPages
        };
        return Partial("_List", ListVm);
    }
}
```

- [ ] **Step 2: Create `dotnet/Pages/Expenses/Index.cshtml`**

```cshtml
@page "/expenses"
@model EstudoHtmx.Pages.Expenses.IndexModel
@{
    Layout = null;
}
@await Html.PartialAsync("_List", Model.ListVm)
```

- [ ] **Step 3: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 4: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Pages/Expenses/Index.cshtml dotnet/Pages/Expenses/Index.cshtml.cs
git commit -m "feat(dotnet): rota get /expenses retorna fragmento da lista paginada"
```

---

### Task 12: Route GET /expenses/form?id=N — form fragment

**Files:**
- Create: `dotnet/Pages/Expenses/Form.cshtml`
- Create: `dotnet/Pages/Expenses/Form.cshtml.cs`

- [ ] **Step 1: Create `dotnet/Pages/Expenses/Form.cshtml.cs`**

```csharp
using System.Globalization;
using EstudoHtmx.Data;
using EstudoHtmx.Pages.Shared;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace EstudoHtmx.Pages.Expenses;

public class FormPageModel : PageModel
{
    private readonly ExpenseRepository _repo;

    public FormPageModel(ExpenseRepository repo)
    {
        _repo = repo;
    }

    public FormModel FormVm { get; private set; } = new();

    public IActionResult OnGet([FromQuery] long? id)
    {
        if (id is null)
        {
            FormVm = new FormModel
            {
                Id = null,
                Nome = "",
                ValorRaw = "",
                Descricao = "",
                TotalContexto = _repo.SumAll(excludingId: null)
            };
            return Partial("_Form", FormVm);
        }

        var e = _repo.GetById(id.Value);
        if (e is null)
        {
            FormVm = new FormModel
            {
                Id = null,
                Nome = "",
                ValorRaw = "",
                Descricao = "",
                TotalContexto = _repo.SumAll(excludingId: null)
            };
            return Partial("_Form", FormVm);
        }

        FormVm = new FormModel
        {
            Id = e.Id,
            Nome = e.Nome,
            ValorRaw = e.Valor.ToString("0.##", CultureInfo.InvariantCulture),
            Descricao = e.Descricao,
            TotalContexto = _repo.SumAll(excludingId: e.Id)
        };
        return Partial("_Form", FormVm);
    }
}
```

- [ ] **Step 2: Create `dotnet/Pages/Expenses/Form.cshtml`**

```cshtml
@page "/expenses/form"
@model EstudoHtmx.Pages.Expenses.FormPageModel
@{
    Layout = null;
}
@await Html.PartialAsync("_Form", Model.FormVm)
```

- [ ] **Step 3: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 4: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Pages/Expenses/Form.cshtml dotnet/Pages/Expenses/Form.cshtml.cs
git commit -m "feat(dotnet): rota get /expenses/form retorna fragmento do form"
```

---

### Task 13: Route GET /expenses/total-context?excluding_id=N — input fragment

**Files:**
- Create: `dotnet/Pages/Expenses/TotalContext.cshtml`
- Create: `dotnet/Pages/Expenses/TotalContext.cshtml.cs`

- [ ] **Step 1: Create `dotnet/Pages/Expenses/TotalContext.cshtml.cs`**

```csharp
using EstudoHtmx.Data;
using EstudoHtmx.Pages.Shared;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace EstudoHtmx.Pages.Expenses;

public class TotalContextPageModel : PageModel
{
    private readonly ExpenseRepository _repo;

    public TotalContextPageModel(ExpenseRepository repo)
    {
        _repo = repo;
    }

    public TotalContextModel Vm { get; private set; } = new();

    public IActionResult OnGet([FromQuery(Name = "excluding_id")] long? excludingId)
    {
        Vm = new TotalContextModel
        {
            ExcludingId = excludingId,
            Total = _repo.SumAll(excludingId: excludingId)
        };
        return Partial("_TotalContext", Vm);
    }
}
```

- [ ] **Step 2: Create `dotnet/Pages/Expenses/TotalContext.cshtml`**

```cshtml
@page "/expenses/total-context"
@model EstudoHtmx.Pages.Expenses.TotalContextPageModel
@{
    Layout = null;
}
@await Html.PartialAsync("_TotalContext", Model.Vm)
```

- [ ] **Step 3: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 4: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Pages/Expenses/TotalContext.cshtml dotnet/Pages/Expenses/TotalContext.cshtml.cs
git commit -m "feat(dotnet): rota get /expenses/total-context retorna input com total"
```

---

### Task 14: Route POST /expenses — submit handler with validation, OOB swap

**Files:**
- Create: `dotnet/Pages/Shared/_FormWithOob.cshtml`
- Create: `dotnet/Pages/Shared/_FormWithOobModel.cs`
- Create: `dotnet/Pages/Expenses/Submit.cshtml`
- Create: `dotnet/Pages/Expenses/Submit.cshtml.cs`

- [ ] **Step 1: Create `dotnet/Pages/Shared/_FormWithOobModel.cs`**

```csharp
namespace EstudoHtmx.Pages.Shared;

public sealed class FormWithOobModel
{
    public FormModel Form { get; set; } = new();
    public ListModel? OobList { get; set; }
}
```

- [ ] **Step 2: Create `dotnet/Pages/Shared/_FormWithOob.cshtml`**

```cshtml
@model EstudoHtmx.Pages.Shared.FormWithOobModel
@await Html.PartialAsync("_Form", Model.Form)
@if (Model.OobList != null)
{
    <table hidden hx-swap-oob="outerHTML:#expense-list">
        @await Html.PartialAsync("_Tbody", Model.OobList)
    </table>
}
```

> Note: htmx evaluates `hx-swap-oob` on any top-level element of the response. Wrapping the `<tbody>` in a hidden `<table>` keeps the HTML parser happy (a bare `<tbody>` is not parseable outside a `<table>`) while `hx-swap-oob="outerHTML:#expense-list"` tells htmx to replace the live `#expense-list` tbody with the parsed tbody from inside the wrapper.

- [ ] **Step 3: Create `dotnet/Pages/Expenses/Submit.cshtml.cs`**

```csharp
using EstudoHtmx.Data;
using EstudoHtmx.Pages.Shared;
using EstudoHtmx.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace EstudoHtmx.Pages.Expenses;

public class SubmitPageModel : PageModel
{
    private readonly ExpenseRepository _repo;

    public SubmitPageModel(ExpenseRepository repo)
    {
        _repo = repo;
    }

    public FormWithOobModel Vm { get; private set; } = new();

    public IActionResult OnPost(
        [FromForm(Name = "id")] string? idRaw,
        [FromForm(Name = "nome")] string? nomeRaw,
        [FromForm(Name = "valor")] string? valorRaw,
        [FromForm(Name = "descricao")] string? descricaoRaw)
    {
        long? id = null;
        if (!string.IsNullOrWhiteSpace(idRaw)
            && long.TryParse(idRaw, out var parsedId)
            && parsedId > 0)
        {
            id = parsedId;
        }

        var nome = (nomeRaw ?? "").Trim();
        var descricao = (descricaoRaw ?? "").Trim();
        var valorInput = valorRaw ?? "";

        string? erroNome = null;
        string? erroValor = null;
        double valorParsed = 0d;

        if (nome.Length == 0)
            erroNome = "nome obrigatorio";

        var parse = ValorParser.Parse(valorInput);
        if (!parse.Ok)
            erroValor = parse.Error ?? "valor invalido";
        else
            valorParsed = parse.Value;

        if (erroNome != null || erroValor != null)
        {
            Vm = new FormWithOobModel
            {
                Form = new FormModel
                {
                    Id = id,
                    Nome = nomeRaw ?? "",
                    ValorRaw = valorInput,
                    Descricao = descricaoRaw ?? "",
                    TotalContexto = _repo.SumAll(excludingId: id),
                    ErroNome = erroNome,
                    ErroValor = erroValor
                },
                OobList = null
            };
            return Partial("_FormWithOob", Vm);
        }

        long affectedId;
        int targetPage;

        if (id.HasValue)
        {
            _repo.Update(id.Value, nome, valorParsed, descricao);
            affectedId = id.Value;
        }
        else
        {
            affectedId = _repo.Insert(nome, valorParsed, descricao);
        }

        targetPage = _repo.PageOfId(affectedId);
        var totalPages = _repo.PageCount();
        if (targetPage > totalPages) targetPage = totalPages;
        if (targetPage < 1) targetPage = 1;

        Vm = new FormWithOobModel
        {
            Form = new FormModel
            {
                Id = null,
                Nome = "",
                ValorRaw = "",
                Descricao = "",
                TotalContexto = _repo.SumAll(excludingId: null)
            },
            OobList = new ListModel
            {
                Rows = _repo.ListPage(targetPage),
                Page = targetPage,
                TotalPages = totalPages
            }
        };

        Response.Headers["HX-Trigger"] = "itemSaved";
        return Partial("_FormWithOob", Vm);
    }
}
```

- [ ] **Step 4: Create `dotnet/Pages/Expenses/Submit.cshtml`**

```cshtml
@page "/expenses"
@model EstudoHtmx.Pages.Expenses.SubmitPageModel
@{
    Layout = null;
}
@await Html.PartialAsync("_FormWithOob", Model.Vm)
```

- [ ] **Step 5: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 6: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Pages/Shared/_FormWithOob.cshtml dotnet/Pages/Shared/_FormWithOobModel.cs dotnet/Pages/Expenses/Submit.cshtml dotnet/Pages/Expenses/Submit.cshtml.cs
git commit -m "feat(dotnet): rota post /expenses com validacao e oob da tabela"
```

---

### Task 15: Route POST /expenses/{id}/zerar — set valor=0 + OOB

**Files:**
- Create: `dotnet/Pages/Expenses/Zerar.cshtml`
- Create: `dotnet/Pages/Expenses/Zerar.cshtml.cs`

- [ ] **Step 1: Create `dotnet/Pages/Expenses/Zerar.cshtml.cs`**

```csharp
using EstudoHtmx.Data;
using EstudoHtmx.Pages.Shared;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace EstudoHtmx.Pages.Expenses;

public class ZerarPageModel : PageModel
{
    private readonly ExpenseRepository _repo;

    public ZerarPageModel(ExpenseRepository repo)
    {
        _repo = repo;
    }

    public FormWithOobModel Vm { get; private set; } = new();

    public IActionResult OnPost(long id, [FromQuery] int page = 1)
    {
        _repo.Zerar(id);

        var totalPages = _repo.PageCount();
        var targetPage = _repo.PageOfId(id);
        if (targetPage > totalPages) targetPage = totalPages;
        if (targetPage < 1) targetPage = 1;

        Vm = new FormWithOobModel
        {
            Form = new FormModel
            {
                Id = null,
                Nome = "",
                ValorRaw = "",
                Descricao = "",
                TotalContexto = _repo.SumAll(excludingId: null)
            },
            OobList = new ListModel
            {
                Rows = _repo.ListPage(targetPage),
                Page = targetPage,
                TotalPages = totalPages
            }
        };

        Response.Headers["HX-Trigger"] = "itemZerado";
        return Partial("_FormWithOob", Vm);
    }
}
```

- [ ] **Step 2: Create `dotnet/Pages/Expenses/Zerar.cshtml`**

```cshtml
@page "/expenses/{id:long}/zerar"
@model EstudoHtmx.Pages.Expenses.ZerarPageModel
@{
    Layout = null;
}
@await Html.PartialAsync("_FormWithOob", Model.Vm)
```

- [ ] **Step 3: Build**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet build dotnet/EstudoHtmx.csproj
```

- [ ] **Step 4: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/Pages/Expenses/Zerar.cshtml dotnet/Pages/Expenses/Zerar.cshtml.cs
git commit -m "feat(dotnet): rota post /expenses/{id}/zerar com swap oob"
```

---

### Task 16: README for the dotnet folder

**Files:**
- Create: `dotnet/README.md`

- [ ] **Step 1: Create `dotnet/README.md`**

```markdown
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
```

- [ ] **Step 2: Commit**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git add dotnet/README.md
git commit -m "docs(dotnet): adiciona readme com instrucoes de instalacao e rodagem"
```

---

### Task 17: Acceptance test checklist (manual)

**Files:**
- None (manual verification only)

- [ ] **Step 1: Start the server**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
dotnet run --project dotnet/EstudoHtmx.csproj
```

Wait for the console to log `Now listening on: http://localhost:5001`.

- [ ] **Step 2: Open the browser**

Open `http://localhost:5001` and confirm: the page renders pico-styled, the form appears above an empty list (`nenhuma despesa cadastrada`), and the total_acumulado input shows `R$ 0,00`.

- [ ] **Step 3: Criar uma despesa com valor literal**

In the form: `nome = aluguel`, `valor = 100`, `descricao = teste`. Click `salvar`. The form resets, the row appears in the table with `valor = R$ 100,00` and `total acumulado = R$ 100,00`.

- [ ] **Step 4: Criar uma despesa com expressao**

In the form: `nome = bonus calc`, `valor = =(-1*(50/2)+10)`, `descricao = expr test`. Click `salvar`. The row appears with `valor = -R$ 15,00` in red and `total acumulado = R$ 85,00`.

- [ ] **Step 5: Editar a despesa criada**

Click the `bi-pencil` icon on the bonus row. The form fills with id, nome, valor. Change `nome` to `bonus editado`, click `salvar`. Confirm the row updates with the new name.

- [ ] **Step 6: Validar erro — nome vazio**

Clear `nome`, type any valid `valor`, click `salvar`. Confirm the form re-renders inline with `<small class="error">nome obrigatorio</small>` and that the inputs keep their typed values.

- [ ] **Step 7: Validar erro — valor=abc**

Type `nome = x`, `valor = abc`, click `salvar`. Confirm an inline error `valor invalido: nao e numero nem expressao` appears.

- [ ] **Step 8: Zerar uma despesa**

Click the `bi-arrow-counterclockwise` icon on the aluguel row, confirm the JS dialog `Zerar este item?`. After confirm, the row remains visible with `valor = R$ 0,00` and the running totals downstream recalculate.

- [ ] **Step 9: Paginacao com 25 registros**

Insert 23 more rows (any sequence). Navigate to page 2 via `>`. Confirm: the running sum on the first row of page 2 equals the sum of valor across page 1's 20 rows. Click `<<` to jump to page 1.

- [ ] **Step 10: Total acumulado atualiza no focusout**

Click in `nome`, then tab out (focusout). Confirm the `total_acumulado` input value reflects the BRL sum of all rows. While editing an existing row, confirm the sum excludes that row's current persisted valor.

- [ ] **Step 11: Persistencia cross-stack (deferred)**

This step is verified after another stack runs. Manual: stop this server, run a sibling stack against `shared/expenses.db`, confirm same rows appear.

- [ ] **Step 12: Stop the server**

`Ctrl+C` in the terminal.

- [ ] **Step 13: Commit a single chore line documenting the run (optional, only if anything was tweaked during testing)**

Run from `C:\Dev\github\estudoHtmx`:

```powershell
git status
```

If nothing changed, skip the commit. If you tweaked styling or copy during manual testing:

```powershell
git add <touched files>
git commit -m "chore(dotnet): ajustes apos checklist de aceitacao"
```
