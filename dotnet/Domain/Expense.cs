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
