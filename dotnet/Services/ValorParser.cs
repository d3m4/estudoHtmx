using System.Data;
using System.Globalization;
using System.Text.RegularExpressions;

namespace EstudoHtmx.Services;

public static class ValorParser
{
    public readonly record struct ParseResult(bool Ok, double Value, string? Error);

    public static ParseResult Parse(string? raw)
    {
        if (string.IsNullOrWhiteSpace(raw))
            return new ParseResult(false, 0d, "valor obrigatorio");

        var trimmed = raw.Trim();

        if (trimmed.StartsWith("="))
        {
            var expr = trimmed[1..].Trim();
            if (expr.Length == 0)
                return new ParseResult(false, 0d, "expressao vazia");

            if (!Regex.IsMatch(expr, @"^[0-9\+\-\*\/\(\)\.\s]+$"))
                return new ParseResult(false, 0d,
                    "expressao invalida: use apenas + - * / ( ) e numeros com ponto decimal");

            try
            {
                var dt = new DataTable();
                var result = dt.Compute(expr, null);
                if (result is null || result is DBNull)
                    return new ParseResult(false, 0d, "expressao invalida");

                var d = Convert.ToDouble(result, CultureInfo.InvariantCulture);
                if (double.IsNaN(d) || double.IsInfinity(d))
                    return new ParseResult(false, 0d, "expressao invalida (divisao por zero?)");

                return new ParseResult(true, d, null);
            }
            catch (DivideByZeroException)
            {
                return new ParseResult(false, 0d, "divisao por zero");
            }
            catch (SyntaxErrorException)
            {
                return new ParseResult(false, 0d, "expressao com sintaxe invalida");
            }
            catch (EvaluateException)
            {
                return new ParseResult(false, 0d, "expressao nao pode ser avaliada");
            }
            catch (OverflowException)
            {
                return new ParseResult(false, 0d, "expressao gerou overflow");
            }
            catch (Exception)
            {
                return new ParseResult(false, 0d, "expressao invalida");
            }
        }

        // numero literal: aceita ponto OU virgula como decimal
        var normalized = trimmed.Replace(",", ".");
        if (!double.TryParse(
                normalized,
                NumberStyles.Float | NumberStyles.AllowLeadingSign,
                CultureInfo.InvariantCulture,
                out var n))
        {
            return new ParseResult(false, 0d, "valor invalido");
        }

        if (double.IsNaN(n) || double.IsInfinity(n))
            return new ParseResult(false, 0d, "valor invalido");

        return new ParseResult(true, n, null);
    }
}
