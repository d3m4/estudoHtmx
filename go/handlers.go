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

// handleListExpenses serve GET /expenses?page=N — fragmento htmx da tabela
// inteira (incluindo paginacao).
func (s *server) handleListExpenses(w http.ResponseWriter, r *http.Request) {
	page := parseIntDefault(r.URL.Query().Get("page"), 1)
	vm, err := s.buildListVM(page, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.renderTemplate(w, "list", vm)
}

// handleFormFragment serve GET /expenses/form?id=N — fragmento htmx do form,
// vazio ou preenchido pra edicao.
func (s *server) handleFormFragment(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	var exp *Expense
	if idStr != "" {
		id := parseInt64(idStr)
		if id > 0 {
			found, err := getExpense(s.db, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			exp = found
		}
	}
	vm, err := s.buildFormVM(exp, "", nil, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.renderTemplate(w, "form", vm)
}

// handleTotalContext serve GET /expenses/total-context?excluding_id=N —
// fragmento htmx do <input> de total_acumulado.
func (s *server) handleTotalContext(w http.ResponseWriter, r *http.Request) {
	excludingId := parseInt64(r.URL.Query().Get("excluding_id"))
	total, err := sumAll(s.db, excludingId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.renderTemplate(w, "total_context", totalCtxVM{
		Total:       total,
		ExcludingId: excludingId,
	})
}

// handleSaveExpense serve POST /expenses. Faz validacao server-side, executa
// insert ou update, e retorna fragmento htmx contendo:
//   - form re-renderizado (zerado em sucesso, com erros em falha)
//   - tbody atualizado via hx-swap-oob (somente em sucesso)
//
// Em sucesso tambem emite o header HX-Trigger: itemSaved.
func (s *server) handleSaveExpense(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form invalido", http.StatusBadRequest)
		return
	}

	id := parseInt64(r.PostFormValue("id"))
	nomeRaw := r.PostFormValue("nome")
	valorRaw := r.PostFormValue("valor")
	descricao := r.PostFormValue("descricao")

	nome := strings.TrimSpace(nomeRaw)
	errs := map[string]string{}

	if nome == "" {
		errs["Nome"] = "nome e obrigatorio"
	}

	valor, valorErr := parseValor(valorRaw)
	if valorErr != nil {
		errs["Valor"] = valorErr.Error()
	}

	if len(errs) > 0 {
		// devolve o form com erros, sem swap-oob da lista.
		exp := &Expense{
			Id:        id,
			Nome:      nomeRaw,
			Descricao: descricao,
		}
		vm, err := s.buildFormVM(exp, valorRaw, errs, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.renderTemplate(w, "form", vm)
		return
	}

	if id > 0 {
		if err := updateExpense(s.db, id, nome, valor, descricao); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		newId, err := insertExpense(s.db, nome, valor, descricao)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = newId
	}

	// pagina atual: derivada do header HX-Current-URL se possivel; default 1.
	page := currentPageFromRequest(r)

	formVM, err := s.buildFormVM(nil, "", nil, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	listVM, err := s.buildListVM(page, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "itemSaved")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// form e o target principal; list vem via hx-swap-oob (template ja inclui
	// o atributo quando .OOB == true).
	if err := s.tmpls.ExecuteTemplate(w, "form", formVM); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.tmpls.ExecuteTemplate(w, "list", listVM); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// currentPageFromRequest tenta extrair a pagina ativa da URL atual do htmx
// (header HX-Current-URL). Fallback: query string "page" ou 1.
func currentPageFromRequest(r *http.Request) int {
	if cur := r.Header.Get("HX-Current-URL"); cur != "" {
		if i := strings.Index(cur, "page="); i >= 0 {
			tail := cur[i+len("page="):]
			end := len(tail)
			for j, c := range tail {
				if c == '&' || c == '#' {
					end = j
					break
				}
			}
			if n, err := strconv.Atoi(tail[:end]); err == nil && n > 0 {
				return n
			}
		}
	}
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			return n
		}
	}
	return 1
}

// handleZerarExpense serve POST /expenses/{id}/zerar. Seta valor=0 e devolve
// form re-renderizado (zerado) + tbody atualizado via hx-swap-oob.
func (s *server) handleZerarExpense(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id := parseInt64(idStr)
	if id <= 0 {
		http.Error(w, "id invalido", http.StatusBadRequest)
		return
	}
	if err := zerarExpense(s.db, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := currentPageFromRequest(r)
	formVM, err := s.buildFormVM(nil, "", nil, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	listVM, err := s.buildListVM(page, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "itemZeroed")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpls.ExecuteTemplate(w, "form", formVM); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.tmpls.ExecuteTemplate(w, "list", listVM); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
