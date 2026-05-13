namespace EstudoHtmx.Pages.Shared;

public sealed class FormWithOobModel
{
    public FormModel Form { get; set; } = new();
    public ListModel? OobList { get; set; }
}
