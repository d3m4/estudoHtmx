import type { FC } from "hono/jsx";
import { formatBRL } from "../format.ts";
import type { ExpenseRow } from "../types.ts";

interface ListProps {
  rows: ExpenseRow[];
  page: number;
  totalPages: number;
  oob?: boolean;
}

export const List: FC<ListProps> = ({ rows, page, totalPages, oob }) => {
  return (
    <article id="expense-table">
      <table role="grid">
        <thead>
          <tr>
            <th scope="col">#</th>
            <th scope="col">nome</th>
            <th scope="col">valor</th>
            <th scope="col">total acumulado</th>
            <th scope="col">descricao</th>
            <th scope="col">acoes</th>
          </tr>
        </thead>
        <tbody
          id="expense-list"
          {...(oob ? { "hx-swap-oob": "true" } : {})}
        >
          {rows.length === 0 ? (
            <tr>
              <td colspan={6}>
                <em class="muted">nenhuma despesa ainda</em>
              </td>
            </tr>
          ) : (
            rows.map((r) => (
              <tr>
                <td>{r.id}</td>
                <td>{r.nome}</td>
                <td class={r.valor < 0 ? "negative" : undefined}>{formatBRL(r.valor)}</td>
                <td class={r.total_acumulado < 0 ? "negative" : undefined}>{formatBRL(r.total_acumulado)}</td>
                <td class="desc" title={r.descricao}>{r.descricao}</td>
                <td>
                  <div class="actions">
                    <button
                      type="button"
                      class="outline"
                      title="editar"
                      hx-get={`/expenses/form?id=${r.id}`}
                      hx-target="#expense-form"
                      hx-swap="outerHTML"
                    >
                      <i class="bi bi-pencil"></i>
                    </button>
                    <button
                      type="button"
                      class="outline secondary"
                      title="zerar"
                      hx-post={`/expenses/${r.id}/zerar`}
                      hx-target="#expense-form"
                      hx-swap="outerHTML"
                      hx-confirm="Zerar este item?"
                    >
                      <i class="bi bi-arrow-counterclockwise"></i>
                    </button>
                  </div>
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
      <Pagination page={page} totalPages={totalPages} />
    </article>
  );
};

const Pagination: FC<{ page: number; totalPages: number }> = ({ page, totalPages }) => {
  const atFirst = page <= 1;
  const atLast = page >= totalPages;
  const mk = (p: number) => `/expenses?page=${p}`;
  return (
    <nav class="pagination" aria-label="paginacao">
      <button type="button" disabled={atFirst} hx-get={mk(1)} hx-target="#expense-table" hx-swap="outerHTML">
        <i class="bi bi-chevron-double-left"></i>
      </button>
      <button type="button" disabled={atFirst} hx-get={mk(page - 1)} hx-target="#expense-table" hx-swap="outerHTML">
        <i class="bi bi-chevron-left"></i>
      </button>
      <span>pagina {page} de {totalPages}</span>
      <button type="button" disabled={atLast} hx-get={mk(page + 1)} hx-target="#expense-table" hx-swap="outerHTML">
        <i class="bi bi-chevron-right"></i>
      </button>
      <button type="button" disabled={atLast} hx-get={mk(totalPages)} hx-target="#expense-table" hx-swap="outerHTML">
        <i class="bi bi-chevron-double-right"></i>
      </button>
    </nav>
  );
};
