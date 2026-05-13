namespace EstudoHtmx.Pages.Shared;

public sealed class FormModel
{
    public long? Id { get; set; }
    public string Nome { get; set; } = "";
    public string ValorRaw { get; set; } = "";
    public string Descricao { get; set; } = "";
    public double TotalContexto { get; set; } = 0d;
    public string? ErroNome { get; set; }
    public string? ErroValor { get; set; }
}
