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
