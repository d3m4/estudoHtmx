package main

import (
	"fmt"
	"math"
	"strings"
)

// brl formata um valor float64 como moeda pt-BR: "R$ 1.234,56", "-R$ 50,00".
// Implementacao manual para nao depender de bibliotecas de locale (o pacote
// stdlib do Go nao tem formatacao de moeda nativa).
func brl(v float64) string {
	negative := v < 0
	abs := math.Abs(v)
	// arredonda para 2 casas
	rounded := math.Round(abs*100) / 100

	// separa parte inteira e centavos como strings
	intPart := int64(rounded)
	cents := int64(math.Round((rounded - float64(intPart)) * 100))
	// corrige casos de arredondamento que estouram pra 100
	if cents == 100 {
		intPart++
		cents = 0
	}

	intStr := groupThousands(fmt.Sprintf("%d", intPart))
	body := fmt.Sprintf("R$ %s,%02d", intStr, cents)
	if negative {
		return "-" + body
	}
	return body
}

// groupThousands insere "." como separador de milhar (pt-BR) numa string
// composta apenas por digitos.
func groupThousands(s string) string {
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		if len(s) > pre {
			b.WriteString(".")
		}
	}
	for i := pre; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteString(".")
		}
	}
	return b.String()
}

// isNegative facilita o uso em templates html para colorir negativos.
func isNegative(v float64) bool {
	return v < 0
}
