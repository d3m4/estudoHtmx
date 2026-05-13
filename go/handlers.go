package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

// --- view models ---

type formVM struct {
	Expense      Expense
	ValorRaw     string
	Errors       map[string]string
	TotalContext totalCtxVM
	OOB          bool
}

type totalCtxVM struct {
	Total       float64
	ExcludingId int64
}

type listVM struct {
	Rows       []ExpenseRow
	Page       int
	TotalPages int
	PrevPage   int
	NextPage   int
	OOB        bool
}

type pageVM struct {
	Form formVM
	List listVM
}

// --- server ---

type server struct {
	db    *sql.DB
	tmpls *template.Template
}

func newServer(db *sql.DB) (*server, error) {
	funcs := template.FuncMap{
		"brl": brl,
	}
	t, err := template.New("").Funcs(funcs).ParseGlob("templates/*.tmpl")
	if err != nil {
		return nil, err
	}
	return &server{db: db, tmpls: t}, nil
}

// --- helpers ---

func (s *server) renderTemplate(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpls.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// buildListVM monta o view-model da listagem para a pagina indicada.
func (s *server) buildListVM(page int, oob bool) (listVM, error) {
	rows, total, err := listExpenses(s.db, page)
	if err != nil {
		return listVM{}, err
	}
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	prev := page - 1
	if prev < 1 {
		prev = 1
	}
	next := page + 1
	if next > totalPages {
		next = totalPages
	}
	return listVM{
		Rows:       rows,
		Page:       page,
		TotalPages: totalPages,
		PrevPage:   prev,
		NextPage:   next,
		OOB:        oob,
	}, nil
}

// buildFormVM monta o view-model do form, opcionalmente preenchido por um
// expense existente. Para form novo passe expense=nil.
func (s *server) buildFormVM(exp *Expense, valorRaw string, errs map[string]string, oob bool) (formVM, error) {
	var e Expense
	var excludingId int64
	if exp != nil {
		e = *exp
		excludingId = exp.Id
	}
	total, err := sumAll(s.db, excludingId)
	if err != nil {
		return formVM{}, err
	}
	if errs == nil {
		errs = map[string]string{}
	}
	// se valorRaw nao foi fornecido, derive do valor numerico (so quando
	// preenchendo form pra edicao).
	if valorRaw == "" && exp != nil {
		valorRaw = strconv.FormatFloat(exp.Valor, 'f', -1, 64)
		// usa virgula como decimal pra refletir o formato pt-BR de entrada.
		valorRaw = strings.Replace(valorRaw, ".", ",", 1)
	}
	return formVM{
		Expense:      e,
		ValorRaw:     valorRaw,
		Errors:       errs,
		TotalContext: totalCtxVM{Total: total, ExcludingId: excludingId},
		OOB:          oob,
	}, nil
}

// parseIntDefault le um inteiro de uma string com fallback.
func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// parseInt64 le um int64 ou retorna 0 se invalido.
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// --- handlers ---

// handleIndex serve a pagina completa em GET /.
func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	listVM, err := s.buildListVM(1, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	formVM, err := s.buildFormVM(nil, "", nil, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.renderTemplate(w, "layout", pageVM{Form: formVM, List: listVM})
}
