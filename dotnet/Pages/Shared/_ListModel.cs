using EstudoHtmx.Domain;

namespace EstudoHtmx.Pages.Shared;

public sealed class ListModel
{
    public IReadOnlyList<ExpenseRow> Rows { get; set; } = Array.Empty<ExpenseRow>();
    public int Page { get; set; } = 1;
    public int TotalPages { get; set; } = 1;
}
