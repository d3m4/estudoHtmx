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
