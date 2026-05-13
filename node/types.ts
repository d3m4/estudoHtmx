export interface Expense {
  id: number;
  nome: string;
  valor: number;
  descricao: string;
  created_at: string;
}

// linha retornada pelo SELECT com window function (inclui running sum)
export interface ExpenseRow extends Expense {
  total_acumulado: number;
}

// dados que vem do form submit
export interface ExpenseFormData {
  id?: number;
  nome: string;
  valor: string; // ainda nao parseado
  descricao: string;
}

// erros de validacao por campo
export interface ExpenseFormErrors {
  nome?: string;
  valor?: string;
}

// payload usado pelos componentes pra re-renderizar form com estado preservado
export interface FormState {
  id?: number;
  nome: string;
  valor: string;
  descricao: string;
  errors: ExpenseFormErrors;
}

export const PAGE_SIZE = 20;
