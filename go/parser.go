package main

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
)

// errInvalidValor e devolvido pelo parser quando a entrada nao pode ser
// interpretada nem como numero nem como expressao.
var errInvalidValor = errors.New("valor invalido")

// parseValor recebe a string crua vinda do form e retorna o numero
// resultante. Regras:
//   - se comeca com "=", trata o restante como expressao aritmetica
//     (operadores +, -, *, /, parens, decimais com ponto). Avaliado via
//     expr-lang em modo sandboxed.
//   - caso contrario, trata como numero literal aceitando ponto OU virgula
//     como separador decimal.
//
// Retorna erro de validacao com mensagem amigavel em portugues para qualquer
// entrada invalida (sintaxe, divisao por zero, NaN, infinito etc).
func parseValor(raw string) (float64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, errInvalidValor
	}

	if strings.HasPrefix(s, "=") {
		return evalExpression(strings.TrimSpace(s[1:]))
	}
	return parseLiteralNumber(s)
}

// parseLiteralNumber aceita ponto OU virgula como separador decimal.
// Rejeita strings que contenham letras ou multiplos separadores.
func parseLiteralNumber(s string) (float64, error) {
	// normaliza virgula -> ponto. Nao aceita milhar (entrada do usuario nao
	// usa "1.234,56" pra digitar).
	normalized := strings.ReplaceAll(s, ",", ".")
	if strings.Count(normalized, ".") > 1 {
		return 0, errInvalidValor
	}
	v, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, errInvalidValor
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, errInvalidValor
	}
	return v, nil
}

// evalExpression compila e executa uma expressao aritmetica usando expr-lang.
// Restringe a sintaxe permitida atraves de validacao previa simples: apenas
// digitos, ponto, espacos e operadores + - * / ( ).
func evalExpression(expression string) (float64, error) {
	if expression == "" {
		return 0, errInvalidValor
	}
	if !allowedExprChars(expression) {
		return 0, errInvalidValor
	}

	program, err := expr.Compile(expression, expr.AsFloat64())
	if err != nil {
		return 0, errInvalidValor
	}
	out, err := expr.Run(program, nil)
	if err != nil {
		// captura divisao por zero, overflow e afins
		return 0, errInvalidValor
	}
	v, ok := out.(float64)
	if !ok {
		return 0, errInvalidValor
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, errInvalidValor
	}
	return v, nil
}

// allowedExprChars garante que a expressao usa apenas o alfabeto permitido:
// digitos, ponto, espaco em branco e operadores aritmeticos basicos.
func allowedExprChars(s string) bool {
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r == '.' || r == '+' || r == '-' || r == '*' || r == '/' || r == '(' || r == ')':
		case r == ' ' || r == '\t':
		default:
			return false
		}
	}
	return true
}
