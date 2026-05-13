import { Parser } from "expr-eval";

// instancia configurada minimamente (sem funcoes built-in que poderiamos remover,
// mas a lib ja nao expõe `eval` do host — sandbox por design)
const parser = new Parser({
  operators: {
    add: true,
    subtract: true,
    multiply: true,
    divide: true,
    power: false,
    factorial: false,
    remainder: false,
    logical: false,
    comparison: false,
    in: false,
    assignment: false,
    concatenate: false,
  },
});

export interface ParseResult {
  ok: boolean;
  value?: number;
  error?: string;
}

/**
 * Parseia o campo `valor` do form.
 * - Se a string comeca com `=`, avalia como expressao (ponto como decimal).
 * - Caso contrario, aceita numero literal com ponto OU virgula como decimal.
 * Retorna {ok:true, value} ou {ok:false, error}.
 */
export function parseValor(raw: string): ParseResult {
  const trimmed = (raw ?? "").trim();
  if (trimmed === "") {
    return { ok: false, error: "informe um valor" };
  }

  if (trimmed.startsWith("=")) {
    const expr = trimmed.slice(1).trim();
    if (expr === "") {
      return { ok: false, error: "expressao vazia apos =" };
    }
    // por convencao da spec, virgula nao e aceita em expressao
    if (expr.includes(",")) {
      return { ok: false, error: "use ponto como decimal em expressoes" };
    }
    try {
      const compiled = parser.parse(expr);
      const value = compiled.evaluate({}); // contexto vazio = sem variaveis
      if (typeof value !== "number" || !Number.isFinite(value)) {
        return { ok: false, error: "expressao nao gerou numero finito" };
      }
      return { ok: true, value };
    } catch (err) {
      const msg = err instanceof Error ? err.message : "erro de sintaxe";
      return { ok: false, error: `expressao invalida: ${msg}` };
    }
  }

  // numero literal: aceita ponto OU virgula como decimal
  // remove separadores de milhar tipo "1.234,56" -> "1234.56"
  // estrategia: se tem virgula, virgula e decimal e ponto e milhar; senao ponto e decimal.
  let normalized: string;
  if (trimmed.includes(",")) {
    normalized = trimmed.replace(/\./g, "").replace(",", ".");
  } else {
    normalized = trimmed;
  }
  const n = Number(normalized);
  if (!Number.isFinite(n)) {
    return { ok: false, error: "valor invalido" };
  }
  return { ok: true, value: n };
}
