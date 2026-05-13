const fmt = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

/** Formata 1234.56 -> "R$ 1.234,56"; -50 -> "-R$ 50,00". */
export function formatBRL(value: number): string {
  return fmt.format(value);
}

export function isNegative(value: number): boolean {
  return value < 0;
}
