import type { FC } from "hono/jsx";
import { formatBRL } from "../format.ts";

interface TotalContextProps {
  excludingId: number | null;
  total: number;
}

export const TotalContext: FC<TotalContextProps> = ({ excludingId, total }) => {
  const param = excludingId != null ? String(excludingId) : "";
  const url = `/expenses/total-context?excluding_id=${param}`;
  return (
    <input
      id="total_acumulado"
      type="text"
      name="total_acumulado"
      readonly
      value={formatBRL(total)}
      hx-get={url}
      hx-trigger="focusout from:#expense-form input delay:300ms, focusout from:#expense-form textarea delay:300ms"
      hx-target="this"
      hx-swap="outerHTML"
    />
  );
};
