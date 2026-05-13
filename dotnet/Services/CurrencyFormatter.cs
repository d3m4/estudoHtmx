using System.Globalization;

namespace EstudoHtmx.Services;

public static class CurrencyFormatter
{
    private static readonly CultureInfo PtBr = new("pt-BR");

    public static string Brl(double value)
    {
        // pt-BR ja gera "R$ 1.234,56" e "-R$ 50,00"
        return value.ToString("C", PtBr);
    }

    public static bool IsNegative(double value) => value < 0d;
}
