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
import type { FormState } from "./types.ts";

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

const port = Number(process.env.PORT ?? 5004);
console.log(`[node] servidor escutando em http://localhost:${port}`);
export default { port, fetch: app.fetch };
