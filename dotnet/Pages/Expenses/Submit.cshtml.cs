using EstudoHtmx.Data;
using EstudoHtmx.Pages.Shared;
using EstudoHtmx.Services;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace EstudoHtmx.Pages.Expenses;

public class SubmitPageModel : PageModel
{
    private readonly ExpenseRepository _repo;

    public SubmitPageModel(ExpenseRepository repo)
    {
        _repo = repo;
    }

    public FormWithOobModel Vm { get; private set; } = new();

    public IActionResult OnPost(
        [FromForm(Name = "id")] string? idRaw,
        [FromForm(Name = "nome")] string? nomeRaw,
        [FromForm(Name = "valor")] string? valorRaw,
        [FromForm(Name = "descricao")] string? descricaoRaw)
    {
        long? id = null;
        if (!string.IsNullOrWhiteSpace(idRaw)
            && long.TryParse(idRaw, out var parsedId)
            && parsedId > 0)
        {
            id = parsedId;
        }

        var nome = (nomeRaw ?? "").Trim();
        var descricao = (descricaoRaw ?? "").Trim();
        var valorInput = valorRaw ?? "";

        string? erroNome = null;
        string? erroValor = null;
        double valorParsed = 0d;

        if (nome.Length == 0)
            erroNome = "nome obrigatorio";

        var parse = ValorParser.Parse(valorInput);
        if (!parse.Ok)
            erroValor = parse.Error ?? "valor invalido";
        else
            valorParsed = parse.Value;

        if (erroNome != null || erroValor != null)
        {
            Vm = new FormWithOobModel
            {
                Form = new FormModel
                {
                    Id = id,
                    Nome = nomeRaw ?? "",
                    ValorRaw = valorInput,
                    Descricao = descricaoRaw ?? "",
                    TotalContexto = _repo.SumAll(excludingId: id),
                    ErroNome = erroNome,
                    ErroValor = erroValor
                },
                OobList = null
            };
            return Partial("_FormWithOob", Vm);
        }

        long affectedId;
        int targetPage;

        if (id.HasValue)
        {
            _repo.Update(id.Value, nome, valorParsed, descricao);
            affectedId = id.Value;
        }
        else
        {
            affectedId = _repo.Insert(nome, valorParsed, descricao);
        }

        targetPage = _repo.PageOfId(affectedId);
        var totalPages = _repo.PageCount();
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

        Response.Headers["HX-Trigger"] = "itemSaved";
        return Partial("_FormWithOob", Vm);
    }
}
