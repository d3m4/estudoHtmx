mod db;
mod expense;
mod format;
mod handlers;
mod parser;

use axum::routing::{get, post};
use axum::Router;
use handlers::AppState;
use std::net::SocketAddr;

#[tokio::main]
async fn main() {
    let db = db::open_and_ensure_schema("../shared/expenses.db", "../shared/schema.sql")
        .expect("falha abrindo banco");
    let state = AppState { db };

    let app = Router::new()
        .route("/", get(handlers::index))
        .route("/expenses", get(handlers::get_expenses).post(handlers::post_expense))
        .route("/expenses/form", get(handlers::get_form))
        .route("/expenses/total-context", get(handlers::get_total_context))
        .route("/expenses/:id/zerar", post(handlers::post_zerar))
        .with_state(state);

    let addr: SocketAddr = "127.0.0.1:5005".parse().unwrap();
    println!("estudoHtmx-rust ouvindo em http://{addr}");
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
