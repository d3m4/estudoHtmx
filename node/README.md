# node — backend hono + jsx + htmx

Implementacao Node.js do estudoHtmx, rodando com **Bun** + **Hono** + **better-sqlite3**. Compartilha `../shared/expenses.db` com os outros 4 stacks.

## requisitos

- [Bun](https://bun.sh) 1.0+ (recomendado, sem build step, TypeScript e JSX nativos)
- Alternativa: Node.js 20+ com `tsx` (veja secao "fallback node.js" abaixo)

## instalacao

A partir desta pasta (`node/`):

```bash
bun install
```

(o `bun.lockb` esta gitignored; o install reinstala tudo do zero baseado no `package.json`.)

## rodar

Modo dev com hot reload:

```bash
bun --hot run server.tsx
```

Modo "producao" (sem hot reload):

```bash
bun run server.tsx
```

A aplicacao sobe em `http://localhost:5004`.

## portas e arquivos

- porta: **5004**
- db compartilhado: `../shared/expenses.db` (criado automaticamente no primeiro start)
- schema: `../shared/schema.sql` (executado a cada start, idempotente)

## estrutura

```
node/
├── server.tsx              # entrypoint hono + rotas
├── db.ts                   # conexao sqlite + ensure schema
├── repository.ts           # CRUD + paginacao + running sum global
├── parser.ts               # parser de expressao (expr-eval, sandboxed)
├── format.ts               # formatador BRL pt-BR
├── types.ts                # tipos de dominio
├── tsconfig.json           # config crucial: jsx + jsxImportSource hono/jsx
├── package.json
└── components/
    ├── Layout.tsx          # html shell + pico + bootstrap icons
    ├── Form.tsx            # form com hint, erros inline, total ajax
    ├── List.tsx            # tabela + paginacao
    └── TotalContext.tsx    # input total_acumulado renderizavel isolado
```

## fallback node.js (sem Bun)

Se preferir Node.js puro:

```bash
npm install
npm install --save-dev tsx
npx tsx server.tsx
```

Adicione ao topo do `server.tsx` se rodar fora do Bun:

```ts
// substituir uso de `import.meta.dir` (Bun-specific) por:
// import { fileURLToPath } from "node:url";
// import { dirname } from "node:path";
// const __dirname = dirname(fileURLToPath(import.meta.url));
```

E adapte `db.ts` para usar `__dirname` no lugar de `import.meta.dir`. Tambem trocar o `export default { port, fetch }` por `serve({ fetch: app.fetch, port })` usando `@hono/node-server`. Para o estudo, **Bun e a recomendacao**.

## observacoes

- O OOB do tbody no POST sempre re-renderiza a pagina **onde o item salvo caiu** (calculada por `pageOfId`). A paginacao visivel pode ficar momentaneamente fora de sync com o tbody quando o item recai numa pagina diferente — basta clicar em qualquer botao de paginacao para resincronizar.
- O running sum (`total_acumulado`) e calculado **globalmente** via window function `SUM(valor) OVER (ORDER BY id)` em uma subquery, antes do `LIMIT/OFFSET`.
- Validacao server-side only (sem JS de validacao no front). Erros chegam como `<small class="error">` inline.
