use crate::db::Db;
use crate::expense::{self, Expense};
use crate::format::format_brl;
use crate::parser;
use askama::Template;
use askama_axum::IntoResponse;
use axum::extract::{Form, Path, Query, State};
use axum::http::HeaderMap;
use axum::response::Html;
use serde::Deserialize;

#[derive(Clone)]
pub struct AppState {
    pub db: Db,
}

#[derive(Template)]
#[template(path = "form.html")]
pub struct FormTpl {
    pub id: Option<i64>,
    pub nome: String,
    pub valor: String,
    pub descricao: String,
    pub total_acumulado_fmt: String,
    pub error_nome: Option<String>,
    pub error_valor: Option<String>,
    pub oob: bool,
}

impl FormTpl {
    pub fn empty(total: f64) -> Self {
        Self {
            id: None,
            nome: String::new(),
            valor: String::new(),
            descricao: String::new(),
            total_acumulado_fmt: format_brl(total),
            error_nome: None,
            error_valor: None,
            oob: false,
        }
    }
}

#[derive(Template)]
#[template(path = "list.html")]
pub struct ListTpl {
    pub items: Vec<Expense>,
    pub page: i64,
    pub total_pages: i64,
    pub oob: bool,
}

#[derive(Template)]
#[template(path = "layout.html", escape = "none")]
pub struct LayoutTpl {
    pub form_html: String,
    pub list_html: String,
}

#[derive(Template)]
#[template(path = "total_context.html")]
pub struct TotalContextTpl {
    pub total_fmt: String,
    pub excluding_id: Option<i64>,
}

pub struct CombinedFragment {
    pub form_html: String,
    pub list_html: String,
}

impl IntoResponse for CombinedFragment {
    fn into_response(self) -> axum::response::Response {
        Html(format!("{}\n{}", self.form_html, self.list_html)).into_response()
    }
}

pub async fn index(State(s): State<AppState>) -> impl IntoResponse {
    let (items, total_count) = expense::list_page(&s.db, 1).unwrap_or((vec![], 0));
    let total_sum = expense::sum_all(&s.db).unwrap_or(0.0);
    let total_pages = expense::total_pages(total_count);

    let form = FormTpl::empty(total_sum);
    let list = ListTpl {
        items,
        page: 1,
        total_pages,
        oob: false,
    };

    let layout = LayoutTpl {
        form_html: form.render().unwrap_or_default(),
        list_html: list.render().unwrap_or_default(),
    };
    Html(layout.render().unwrap_or_default())
}

#[derive(Deserialize)]
pub struct PageQuery {
    pub page: Option<i64>,
}

pub async fn get_expenses(
    State(s): State<AppState>,
    Query(q): Query<PageQuery>,
) -> impl IntoResponse {
    let page = q.page.unwrap_or(1).max(1);
    let (items, total_count) = expense::list_page(&s.db, page).unwrap_or((vec![], 0));
    let total_pages = expense::total_pages(total_count);
    let effective_page = if page > total_pages { total_pages } else { page };

    let (items, _total) = if effective_page != page {
        expense::list_page(&s.db, effective_page).unwrap_or((items, total_count))
    } else {
        (items, total_count)
    };

    let list = ListTpl {
        items,
        page: effective_page,
        total_pages,
        oob: false,
    };
    Html(list.render().unwrap_or_default())
}

#[derive(Deserialize)]
pub struct FormQuery {
    pub id: Option<i64>,
}

pub async fn get_form(
    State(s): State<AppState>,
    Query(q): Query<FormQuery>,
) -> impl IntoResponse {
    let total = match q.id {
        Some(id) => expense::sum_excluding(&s.db, id).unwrap_or(0.0),
        None => expense::sum_all(&s.db).unwrap_or(0.0),
    };

    let form = if let Some(id) = q.id {
        match expense::get_by_id(&s.db, id) {
            Ok(Some(e)) => FormTpl {
                id: Some(e.id),
                nome: e.nome,
                valor: format!("{}", e.valor),
                descricao: e.descricao,
                total_acumulado_fmt: format_brl(total),
                error_nome: None,
                error_valor: None,
                oob: false,
            },
            _ => FormTpl::empty(total),
        }
    } else {
        FormTpl::empty(total)
    };

    Html(form.render().unwrap_or_default())
}

#[derive(Deserialize)]
pub struct TotalContextQuery {
    pub excluding_id: Option<i64>,
}

pub async fn get_total_context(
    State(s): State<AppState>,
    Query(q): Query<TotalContextQuery>,
) -> impl IntoResponse {
    let total = match q.excluding_id {
        Some(id) => expense::sum_excluding(&s.db, id).unwrap_or(0.0),
        None => expense::sum_all(&s.db).unwrap_or(0.0),
    };
    let tpl = TotalContextTpl {
        total_fmt: format_brl(total),
        excluding_id: q.excluding_id,
    };
    Html(tpl.render().unwrap_or_default())
}

#[derive(Deserialize)]
pub struct ExpenseForm {
    pub id: Option<String>,
    pub nome: Option<String>,
    pub valor: Option<String>,
    pub descricao: Option<String>,
}

pub async fn post_expense(
    State(s): State<AppState>,
    _headers: HeaderMap,
    Form(f): Form<ExpenseForm>,
) -> impl IntoResponse {
    let id_opt: Option<i64> = f
        .id
        .as_deref()
        .and_then(|s| s.trim().parse::<i64>().ok())
        .filter(|n| *n > 0);

    let nome_raw = f.nome.unwrap_or_default();
    let valor_raw = f.valor.unwrap_or_default();
    let descricao = f.descricao.unwrap_or_default();

    let nome = nome_raw.trim().to_string();
    let mut error_nome = None;
    let mut error_valor = None;

    if nome.is_empty() {
        error_nome = Some("nome obrigatorio".to_string());
    }

    let parsed_valor = match parser::parse_valor(&valor_raw) {
        Ok(v) => Some(v),
        Err(_e) => {
            error_valor = Some("valor invalido".to_string());
            None
        }
    };

    if error_nome.is_some() || error_valor.is_some() {
        let total = match id_opt {
            Some(id) => expense::sum_excluding(&s.db, id).unwrap_or(0.0),
            None => expense::sum_all(&s.db).unwrap_or(0.0),
        };
        let form = FormTpl {
            id: id_opt,
            nome: nome_raw,
            valor: valor_raw,
            descricao,
            total_acumulado_fmt: format_brl(total),
            error_nome,
            error_valor,
            oob: false,
        };
        return Html(form.render().unwrap_or_default()).into_response();
    }

    let valor = parsed_valor.unwrap();

    match id_opt {
        Some(id) => {
            let _ = expense::update(&s.db, id, &nome, valor, &descricao);
        }
        None => {
            let _ = expense::insert(&s.db, &nome, valor, &descricao);
        }
    }

    let total_sum = expense::sum_all(&s.db).unwrap_or(0.0);
    let (items, total_count) = expense::list_page(&s.db, 1).unwrap_or((vec![], 0));
    let total_pages = expense::total_pages(total_count);

    let mut form = FormTpl::empty(total_sum);
    form.oob = true;

    let list = ListTpl {
        items,
        page: 1,
        total_pages,
        oob: true,
    };

    CombinedFragment {
        form_html: form.render().unwrap_or_default(),
        list_html: list.render().unwrap_or_default(),
    }
    .into_response()
}

pub async fn post_zerar(
    State(s): State<AppState>,
    Path(id): Path<i64>,
) -> impl IntoResponse {
    let _ = expense::zerar(&s.db, id);

    let total_sum = expense::sum_all(&s.db).unwrap_or(0.0);
    let (items, total_count) = expense::list_page(&s.db, 1).unwrap_or((vec![], 0));
    let total_pages = expense::total_pages(total_count);

    let mut form = FormTpl::empty(total_sum);
    form.oob = true;

    let list = ListTpl {
        items,
        page: 1,
        total_pages,
        oob: true,
    };

    CombinedFragment {
        form_html: form.render().unwrap_or_default(),
        list_html: list.render().unwrap_or_default(),
    }
    .into_response()
}
