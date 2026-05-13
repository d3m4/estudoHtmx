import { Hono } from "hono";
import { Layout } from "./components/Layout.tsx";
import { Form, emptyFormState } from "./components/Form.tsx";
import { List } from "./components/List.tsx";
import { TotalContext } from "./components/TotalContext.tsx";
import {
  listExpenses,
  getExpenseById,
  insertExpense,
  updateExpense,
  zerarExpense,
  sumOfAll,
  sumExcluding,
  pageOfId,
} from "./repository.ts";
import { parseValor } from "./parser.ts";
import { formatBRL } from "./format.ts";
import type { ExpenseRow, FormState } from "./types.ts";

function TbodyOob({ rows }: { rows: ExpenseRow[] }) {
  return (
    <tbody id="expense-list" hx-swap-oob="true">
      {rows.length === 0 ? (
        <tr><td colspan={6}><em class="muted">nenhuma despesa ainda</em></td></tr>
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
                <button type="button" class="outline" title="editar"
                  hx-get={`/expenses/form?id=${r.id}`} hx-target="#expense-form" hx-swap="outerHTML">
                  <i class="bi bi-pencil"></i>
                </button>
                <button type="button" class="outline secondary" title="zerar"
                  hx-post={`/expenses/${r.id}/zerar`} hx-target="#expense-form" hx-swap="outerHTML"
                  hx-confirm="Zerar este item?">
                  <i class="bi bi-arrow-counterclockwise"></i>
                </button>
              </div>
            </td>
          </tr>
        ))
      )}
    </tbody>
  );
}

const app = new Hono();

// GET /  -> pagina completa
app.get("/", (c) => {
  const { rows, page, totalPages } = listExpenses(1);
  return c.html(
    <Layout>
      <Form state={emptyFormState()} totalContext={sumOfAll()} />
      <List rows={rows} page={page} totalPages={totalPages} />
    </Layout>
  );
});

// GET /expenses?page=N  -> fragmento <article id="expense-table">
app.get("/expenses", (c) => {
  const pageParam = c.req.query("page");
  const page = pageParam ? parseInt(pageParam, 10) : 1;
  const { rows, page: safePage, totalPages } = listExpenses(page);
  return c.html(<List rows={rows} page={safePage} totalPages={totalPages} />);
});

// GET /expenses/form?id=N  -> form preenchido (ou vazio se sem id)
app.get("/expenses/form", (c) => {
  const idParam = c.req.query("id");
  const id = idParam ? parseInt(idParam, 10) : NaN;

  if (Number.isFinite(id)) {
    const exp = getExpenseById(id);
    if (exp) {
      const state: FormState = {
        id: exp.id,
        nome: exp.nome,
        valor: String(exp.valor),
        descricao: exp.descricao,
        errors: {},
      };
      return c.html(<Form state={state} totalContext={sumExcluding(exp.id)} />);
    }
  }
  // sem id ou id invalido -> form vazio
  return c.html(<Form state={emptyFormState()} totalContext={sumOfAll()} />);
});

// GET /expenses/total-context?excluding_id=N  -> <input id="total_acumulado">
app.get("/expenses/total-context", (c) => {
  const raw = c.req.query("excluding_id");
  const id = raw && raw !== "" ? parseInt(raw, 10) : NaN;
  if (Number.isFinite(id)) {
    return c.html(<TotalContext excludingId={id} total={sumExcluding(id)} />);
  }
  return c.html(<TotalContext excludingId={null} total={sumOfAll()} />);
});

// POST /expenses  -> valida, persiste, retorna <form> + <tbody> oob
app.post("/expenses", async (c) => {
  const body = await c.req.parseBody();
  const rawId = (body.id as string | undefined) ?? "";
  const nome = ((body.nome as string | undefined) ?? "").trim();
  const valorRaw = ((body.valor as string | undefined) ?? "").trim();
  const descricao = (body.descricao as string | undefined) ?? "";

  const errors: FormState["errors"] = {};
  if (nome === "") {
    errors.nome = "nome obrigatorio";
  }
  const valorResult = parseValor(valorRaw);
  if (!valorResult.ok) {
    errors.valor = valorResult.error;
  }

  const idNum = rawId !== "" ? parseInt(rawId, 10) : NaN;
  const editing = Number.isFinite(idNum);

  // se invalido: re-renderiza form com erros, mantem inputs preenchidos.
  // tbody nao precisa de OOB porque os dados nao mudaram.
  if (Object.keys(errors).length > 0) {
    const state: FormState = {
      id: editing ? idNum : undefined,
      nome,
      valor: valorRaw,
      descricao,
      errors,
    };
    const totalCtx = editing ? sumExcluding(idNum) : sumOfAll();
    return c.html(<Form state={state} totalContext={totalCtx} />);
  }

  // valido: insert ou update
  let savedId: number;
  if (editing) {
    updateExpense(idNum, nome, valorResult.value!, descricao);
    savedId = idNum;
  } else {
    savedId = insertExpense(nome, valorResult.value!, descricao);
  }

  // descobrir em qual pagina o item caiu pra re-renderizar o tbody correto
  const targetPage = pageOfId(savedId);
  const { rows } = listExpenses(targetPage);

  c.header("HX-Trigger", "itemSaved");
  return c.html(
    <>
      <Form state={emptyFormState()} totalContext={sumOfAll()} />
      <TbodyOob rows={rows} />
    </>
  );
});

const port = Number(process.env.PORT ?? 5004);
console.log(`[node] servidor escutando em http://localhost:${port}`);
export default { port, fetch: app.fetch };
