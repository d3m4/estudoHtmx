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
