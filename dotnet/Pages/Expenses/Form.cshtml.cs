using System.Globalization;
using EstudoHtmx.Data;
using EstudoHtmx.Pages.Shared;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace EstudoHtmx.Pages.Expenses;

public class FormPageModel : PageModel
{
    private readonly ExpenseRepository _repo;

    public FormPageModel(ExpenseRepository repo)
    {
        _repo = repo;
    }

    public FormModel FormVm { get; private set; } = new();

    public IActionResult OnGet([FromQuery] long? id)
    {
        if (id is null)
        {
            FormVm = new FormModel
            {
                Id = null,
                Nome = "",
                ValorRaw = "",
                Descricao = "",
                TotalContexto = _repo.SumAll(excludingId: null)
            };
            return Partial("_Form", FormVm);
        }

        var e = _repo.GetById(id.Value);
        if (e is null)
        {
            FormVm = new FormModel
            {
                Id = null,
                Nome = "",
                ValorRaw = "",
                Descricao = "",
                TotalContexto = _repo.SumAll(excludingId: null)
            };
            return Partial("_Form", FormVm);
        }

        FormVm = new FormModel
        {
            Id = e.Id,
            Nome = e.Nome,
            ValorRaw = e.Valor.ToString("0.##", CultureInfo.InvariantCulture),
            Descricao = e.Descricao,
            TotalContexto = _repo.SumAll(excludingId: e.Id)
        };
        return Partial("_Form", FormVm);
    }
}
