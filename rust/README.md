# estudoHtmx — rust (axum + askama)

backend rust do estudo. compartilha `../shared/expenses.db` com os outros 4 stacks.

## requisitos

- rust 1.75+ (edition 2021)
- cargo

## install

a partir da raiz do repositorio:

```powershell
cd rust
cargo build
```

## run

```powershell
cd rust
cargo run
```

ou release:

```powershell
cargo run --release
```

opcional, hot reload:

```powershell
cargo install cargo-watch
cargo watch -x run
```

a aplicacao sobe em **http://localhost:5005**.

## estrutura

```
rust/
├── Cargo.toml
├── src/
│   ├── main.rs        # entrypoint + router axum
│   ├── db.rs          # connection + ensure schema
│   ├── expense.rs     # struct + repository
│   ├── parser.rs      # evalexpr wrapper
│   ├── format.rs      # BRL pt-BR
│   └── handlers.rs    # axum handlers + askama templates
└── templates/
    ├── layout.html
    ├── form.html
    ├── list.html
    └── total_context.html
```

## banco

ao subir, le `../shared/schema.sql` e aplica idempotentemente em `../shared/expenses.db`. WAL e foreign keys ligados.
