package servicos

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"tsure/apps/web/internal/auth"
	"tsure/apps/web/internal/middleware"
	"tsure/apps/web/internal/render"
)

// Handler agrupa dependencias HTTP para /servicos.
type Handler struct {
	Store     *Store
	Templates render.Executor
}

// NewHandler cria um Handler.
func NewHandler(store *Store, tmpl render.Executor) *Handler {
	return &Handler{Store: store, Templates: tmpl}
}

// PageData e o ViewModel da pagina /servicos.
type PageData struct {
	Title           string
	User            auth.User
	CSRFField       template.HTML
	Servico         Servico
	IsNew           bool
	Servicos        []Servico
	Classificacoes  []Classificacao
	Query           string
	Error           string
}

// FormData e o subset usado nos parciais de formulario.
type FormData struct {
	CSRFField      template.HTML
	Servico        Servico
	IsNew          bool
	Error          string
	Classificacoes []Classificacao
}

// RowsData wrappa a tabela.
type RowsData struct {
	Servicos []Servico
	Query    string
}

// ServeIndex: GET /servicos (pagina) e POST /servicos (criar).
func (h *Handler) ServeIndex(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.renderPage(w, r, "", Servico{Ativo: true}, true, "")
	case http.MethodPost:
		h.handleCreate(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// ServeByID despacha /servicos/{id}, /servicos/{id}/edit, /servicos/{id}/delete.
func (h *Handler) ServeByID(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/servicos/")
	rest = strings.Trim(rest, "/")
	if rest == "" || rest == "rows" || rest == "new" {
		http.NotFound(w, r)
		return
	}
	parts := strings.SplitN(rest, "/", 2)
	id, err := uuid.Parse(parts[0])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sub := ""
	if len(parts) > 1 {
		sub = parts[1]
	}

	switch {
	case sub == "edit" && r.Method == http.MethodGet:
		h.handleEdit(w, r, id)
	case sub == "delete" && r.Method == http.MethodPost:
		h.handleDelete(w, r, id)
	case sub == "" && r.Method == http.MethodPost:
		h.handleUpdate(w, r, id)
	case sub == "" && r.Method == http.MethodGet:
		sv, err := h.Store.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			h.serverError(w, err)
			return
		}
		h.renderPage(w, r, "", sv, false, "")
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// ServeRows: GET /servicos/rows?q=  HTMX tbody refresh.
func (h *Handler) ServeRows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("q")
	items, err := h.Store.List(r.Context(), q)
	if err != nil {
		h.serverError(w, err)
		return
	}
	h.render(w, "servicos-rows", RowsData{Servicos: items, Query: q})
}

// ServeNew: GET /servicos/new  formulario vazio (HTMX).
func (h *Handler) ServeNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	h.renderForm(w, r, Servico{Ativo: true, UnidadePadrao: "DIARIA"}, true, "")
}

func (h *Handler) handleEdit(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	sv, err := h.Store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.serverError(w, err)
		return
	}
	h.renderForm(w, r, sv, false, "")
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	sv, err := parseForm(r)
	if err != nil {
		h.renderForm(w, r, sv, true, err.Error())
		return
	}
	id, err := h.Store.Create(r.Context(), sv)
	if err != nil {
		h.renderForm(w, r, sv, true, h.userMessage(err))
		return
	}
	sv.ID = id
	h.respondAfterWrite(w, r, sv, false)
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	sv, err := parseForm(r)
	if err != nil {
		sv.ID = id
		h.renderForm(w, r, sv, false, err.Error())
		return
	}
	if err := h.Store.Update(r.Context(), id, sv); err != nil {
		sv.ID = id
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.renderForm(w, r, sv, false, h.userMessage(err))
		return
	}
	sv.ID = id
	h.respondAfterWrite(w, r, sv, false)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	if err := h.Store.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.serverError(w, err)
		return
	}
	h.respondAfterWrite(w, r, Servico{Ativo: true, UnidadePadrao: "DIARIA"}, true)
}

func (h *Handler) respondAfterWrite(w http.ResponseWriter, r *http.Request, sv Servico, isNew bool) {
	items, err := h.Store.List(r.Context(), "")
	if err != nil {
		h.serverError(w, err)
		return
	}
	classes, err := h.Store.ListClassificacoes(r.Context())
	if err != nil {
		h.serverError(w, err)
		return
	}
	data := struct {
		FormData
		Servicos []Servico
	}{
		FormData: FormData{
			CSRFField:      middleware.CSRFTemplateTag(r).(template.HTML),
			Servico:        sv,
			IsNew:          isNew,
			Classificacoes: classes,
		},
		Servicos: items,
	}
	h.render(w, "servicos-form-with-rows", data)
}

func (h *Handler) renderPage(w http.ResponseWriter, r *http.Request, query string, sv Servico, isNew bool, errMsg string) {
	items, err := h.Store.List(r.Context(), query)
	if err != nil {
		h.serverError(w, err)
		return
	}
	classes, err := h.Store.ListClassificacoes(r.Context())
	if err != nil {
		h.serverError(w, err)
		return
	}
	user, _ := middleware.UserFromContext(r.Context())
	data := PageData{
		Title:          "Servicos  tsure",
		User:           user,
		CSRFField:      middleware.CSRFTemplateTag(r).(template.HTML),
		Servico:        sv,
		IsNew:          isNew,
		Servicos:       items,
		Classificacoes: classes,
		Query:          query,
		Error:          errMsg,
	}
	h.render(w, "servicos", data)
}

func (h *Handler) renderForm(w http.ResponseWriter, r *http.Request, sv Servico, isNew bool, errMsg string) {
	classes, err := h.Store.ListClassificacoes(r.Context())
	if err != nil {
		h.serverError(w, err)
		return
	}
	data := FormData{
		CSRFField:      middleware.CSRFTemplateTag(r).(template.HTML),
		Servico:        sv,
		IsNew:          isNew,
		Error:          errMsg,
		Classificacoes: classes,
	}
	h.render(w, "servicos-form", data)
}

func (h *Handler) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.Templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "render: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) serverError(w http.ResponseWriter, err error) {
	http.Error(w, "erro interno: "+err.Error(), http.StatusInternalServerError)
}

func (h *Handler) userMessage(err error) string {
	switch {
	case errors.Is(err, ErrDuplicate):
		return "Codigo de servico ja cadastrado."
	case errors.Is(err, ErrInvalidInput):
		return strings.TrimPrefix(err.Error(), "dados invalidos: ")
	default:
		return "Nao foi possivel salvar: " + err.Error()
	}
}

func parseForm(r *http.Request) (Servico, error) {
	if err := r.ParseForm(); err != nil {
		return Servico{}, errors.New("formulario invalido")
	}
	get := func(k string) string { return strings.TrimSpace(r.FormValue(k)) }
	valor, _ := strconv.ParseFloat(strings.Replace(get("valor_referencia"), ",", ".", 1), 64)

	sv := Servico{
		Codigo:          get("codigo"),
		Descricao:       get("descricao"),
		UnidadePadrao:   get("unidade_padrao"),
		ValorReferencia: valor,
		Ativo:           get("ativo") == "" || get("ativo") == "on" || get("ativo") == "true" || get("ativo") == "1",
	}
	if raw := get("classificacao_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			sv.ClassificacaoID = &id
		}
	}
	return sv, nil
}
