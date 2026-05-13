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

const port = Number(process.env.PORT ?? 5004);
console.log(`[node] servidor escutando em http://localhost:${port}`);
export default { port, fetch: app.fetch };
