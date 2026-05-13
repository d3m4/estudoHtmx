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
