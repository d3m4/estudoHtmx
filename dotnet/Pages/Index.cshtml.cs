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
