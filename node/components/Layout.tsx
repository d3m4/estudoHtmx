import type { FC, PropsWithChildren } from "hono/jsx";

export const Layout: FC<PropsWithChildren> = ({ children }) => {
  return (
    <html lang="pt-BR">
      <head>
        <meta charset="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>estudoHtmx — node</title>
        <link
          rel="stylesheet"
          href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css"
        />
        <link
          rel="stylesheet"
          href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.3/font/bootstrap-icons.min.css"
        />
        <script
          src="https://unpkg.com/htmx.org@1.9.12"
          integrity="sha384-ujb1lZYygJmzgSwoxRggbCHcjc0rB2XoQrxeTUQyRjrOnlCoYta87iKBWq3EsdM2"
          crossorigin="anonymous"
        ></script>
        <style>{`
          .error { color: var(--pico-color-red-500); display: block; margin-top: 0.25rem; }
          .negative { color: var(--pico-color-red-500); }
          .actions { display: flex; gap: 0.5rem; }
          .actions button { padding: 0.25rem 0.5rem; }
          .pagination { display: flex; gap: 0.5rem; align-items: center; justify-content: center; margin-top: 1rem; }
          .pagination button:disabled { opacity: 0.4; cursor: not-allowed; }
          .muted { color: var(--pico-muted-color); font-size: 0.85rem; }
          td.desc { max-width: 240px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
        `}</style>
      </head>
      <body>
        <main class="container">
          <hgroup>
            <h1>despesas</h1>
            <p class="muted">node + hono + htmx — porta 5004</p>
          </hgroup>
          {children}
        </main>
      </body>
    </html>
  );
};
