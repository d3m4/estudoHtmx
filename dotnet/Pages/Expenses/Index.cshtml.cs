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
