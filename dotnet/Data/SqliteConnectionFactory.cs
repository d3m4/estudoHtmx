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
