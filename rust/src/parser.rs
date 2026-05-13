use evalexpr::{eval, Value};

#[derive(Debug)]
pub enum ParseError {
    Empty,
    InvalidNumber,
    InvalidExpression(String),
    DivisionByZero,
    NotFinite,
}

impl std::fmt::Display for ParseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ParseError::Empty => write!(f, "valor obrigatorio"),
            ParseError::InvalidNumber => write!(f, "nao e um numero valido"),
            ParseError::InvalidExpression(m) => write!(f, "expressao invalida: {m}"),
            ParseError::DivisionByZero => write!(f, "divisao por zero"),
            ParseError::NotFinite => write!(f, "resultado nao finito"),
        }
    }
}

pub fn parse_valor(input: &str) -> Result<f64, ParseError> {
    let s = input.trim();
    if s.is_empty() {
        return Err(ParseError::Empty);
    }

    if let Some(expr) = s.strip_prefix('=') {
        let expr = expr.trim();
        if expr.is_empty() {
            return Err(ParseError::Empty);
        }
        if expr.contains(',') {
            return Err(ParseError::InvalidExpression(
                "use ponto como decimal em expressoes".to_string(),
            ));
        }
        for c in expr.chars() {
            let ok = c.is_ascii_digit()
                || matches!(c, '+' | '-' | '*' | '/' | '(' | ')' | '.' | ' ' | '\t');
            if !ok {
                return Err(ParseError::InvalidExpression(format!(
                    "caractere nao permitido: '{c}'"
                )));
            }
        }
        match eval(expr) {
            Ok(Value::Float(n)) => check_finite(n),
            Ok(Value::Int(n)) => check_finite(n as f64),
            Ok(other) => Err(ParseError::InvalidExpression(format!(
                "tipo nao numerico: {other:?}"
            ))),
            Err(e) => {
                let msg = e.to_string();
                if msg.to_lowercase().contains("division") {
                    Err(ParseError::DivisionByZero)
                } else {
                    Err(ParseError::InvalidExpression(msg))
                }
            }
        }
    } else {
        let normalized = s.replace(',', ".");
        match normalized.parse::<f64>() {
            Ok(n) => check_finite(n),
            Err(_) => Err(ParseError::InvalidNumber),
        }
    }
}

fn check_finite(n: f64) -> Result<f64, ParseError> {
    if n.is_finite() {
        Ok(n)
    } else {
        Err(ParseError::NotFinite)
    }
}
