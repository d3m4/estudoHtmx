namespace EstudoHtmx.Data;

public sealed class SqliteOptions
{
    public string SqliteDbPath { get; set; } = "../shared/expenses.db";
    public string SqliteSchemaPath { get; set; } = "../shared/schema.sql";
}
