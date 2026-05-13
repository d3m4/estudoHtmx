import type { FC } from "hono/jsx";
import { formatBRL } from "../format.ts";
import type { FormState } from "../types.ts";

interface FormProps {
  state: FormState;
  totalContext: number;
  oob?: boolean; // marca para hx-swap-oob na resposta de POST
}

export const Form: FC<FormProps> = ({ state, totalContext, oob }) => {
  const idValue = state.id != null ? String(state.id) : "";
  const totalUrl = `/expenses/total-context?excluding_id=${idValue}`;
  const editing = state.id != null;

  return (
    <form
      id="expense-form"
      hx-post="/expenses"
      hx-target="#expense-form"
      hx-swap="outerHTML"
      {...(oob ? { "hx-swap-oob": "true" } : {})}
    >
      <input type="hidden" name="id" value={idValue} />

      <label>
        nome
        <input
          type="text"
          name="nome"
          required
          value={state.nome}
          aria-invalid={state.errors.nome ? "true" : undefined}
        />
        {state.errors.nome ? (
          <small class="error">{state.errors.nome}</small>
        ) : null}
      </label>

      <label>
        valor
        <input
          type="text"
          name="valor"
          required
          value={state.valor}
          aria-invalid={state.errors.valor ? "true" : undefined}
        />
        <small class="hint">ex: 100, -50, 1234,56 ou =(50/2)+10</small>
        {state.errors.valor ? (
          <small class="error">{state.errors.valor}</small>
        ) : null}
      </label>

      <label>
        total acumulado
        <input
          id="total_acumulado"
          type="text"
          name="total_acumulado"
          readonly
          value={formatBRL(totalContext)}
          hx-get={totalUrl}
          hx-trigger="focusout from:#expense-form input delay:300ms, focusout from:#expense-form textarea delay:300ms"
          hx-target="this"
          hx-swap="outerHTML"
        />
      </label>

      <label>
        descricao
        <textarea name="descricao" rows={2}>{state.descricao}</textarea>
      </label>

      <div role="group">
        <button type="submit">salvar</button>
        <button
          type="button"
          class="secondary"
          hx-get="/expenses/form"
          hx-target="#expense-form"
          hx-swap="outerHTML"
        >
          {editing ? "cancelar" : "limpar"}
        </button>
      </div>
    </form>
  );
};

export function emptyFormState(): FormState {
  return { id: undefined, nome: "", valor: "", descricao: "", errors: {} };
}
