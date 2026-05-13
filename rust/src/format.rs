/// wrapper para Askama (templates passam &f64 nos campos)
pub fn brl(n: &f64) -> String {
    format_brl(*n)
}

pub fn format_brl(n: f64) -> String {
    let negative = n < 0.0;
    let abs = n.abs();
    let cents = (abs * 100.0).round() as i64;
    let int_part = cents / 100;
    let frac_part = cents % 100;

    let int_str = int_part.to_string();
    let mut with_dots = String::new();
    let chars: Vec<char> = int_str.chars().collect();
    for (i, c) in chars.iter().enumerate() {
        if i > 0 && (chars.len() - i) % 3 == 0 {
            with_dots.push('.');
        }
        with_dots.push(*c);
    }

    let body = format!("R$ {with_dots},{frac_part:02}");
    if negative {
        format!("-{body}")
    } else {
        body
    }
}
