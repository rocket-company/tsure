package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"tsure/apps/web/internal/clientes"
)

// ClientesHandler expoe o CRUD de clientes como JSON REST.
type ClientesHandler struct {
	Store *clientes.Store
}

type clienteJSON struct {
	ID              string `json:"id"`
	Tipo            string `json:"tipo"`
	NomeRazaoSocial string `json:"nome_razao_social"`
	Documento       string `json:"documento"`
	Email           string `json:"email"`
	TelefoneFixo    string `json:"telefone_fixo"`
	TelefoneCelular string `json:"telefone_celular"`
	ContatoCliente  string `json:"contato_cliente"`
	Logradouro      string `json:"logradouro"`
	Numero          string `json:"numero"`
	Complemento     string `json:"complemento"`
	Bairro          string `json:"bairro"`
	Cidade          string `json:"cidade"`
	UF              string `json:"uf"`
	CEP             string `json:"cep"`
	Bloqueado       bool   `json:"bloqueado"`
	MotivoBloqueio  string `json:"motivo_bloqueio"`
	Observacoes     string `json:"observacoes"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func toClienteJSON(c clientes.Cliente) clienteJSON {
	return clienteJSON{
		ID:              c.ID.String(),
		Tipo:            c.Tipo,
		NomeRazaoSocial: c.NomeRazaoSocial,
		Documento:       c.Documento,
		Email:           c.Email,
		TelefoneFixo:    c.TelefoneFixo,
		TelefoneCelular: c.TelefoneCelular,
		ContatoCliente:  c.ContatoCliente,
		Logradouro:      c.Logradouro,
		Numero:          c.Numero,
		Complemento:     c.Complemento,
		Bairro:          c.Bairro,
		Cidade:          c.Cidade,
		UF:              c.UF,
		CEP:             c.CEP,
		Bloqueado:       c.Bloqueado,
		MotivoBloqueio:  c.MotivoBloqueio,
		Observacoes:     c.Observacoes,
		CreatedAt:       c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       c.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

type clienteInput struct {
	Tipo            string `json:"tipo"`
	NomeRazaoSocial string `json:"nome_razao_social"`
	Documento       string `json:"documento"`
	Email           string `json:"email"`
	TelefoneFixo    string `json:"telefone_fixo"`
	TelefoneCelular string `json:"telefone_celular"`
	ContatoCliente  string `json:"contato_cliente"`
	Logradouro      string `json:"logradouro"`
	Numero          string `json:"numero"`
	Complemento     string `json:"complemento"`
	Bairro          string `json:"bairro"`
	Cidade          string `json:"cidade"`
	UF              string `json:"uf"`
	CEP             string `json:"cep"`
	Bloqueado       bool   `json:"bloqueado"`
	MotivoBloqueio  string `json:"motivo_bloqueio"`
	Observacoes     string `json:"observacoes"`
}

func (inp clienteInput) toModel() clientes.Cliente {
	return clientes.Cliente{
		Tipo:            inp.Tipo,
		NomeRazaoSocial: inp.NomeRazaoSocial,
		Documento:       inp.Documento,
		Email:           inp.Email,
		TelefoneFixo:    inp.TelefoneFixo,
		TelefoneCelular: inp.TelefoneCelular,
		ContatoCliente:  inp.ContatoCliente,
		Logradouro:      inp.Logradouro,
		Numero:          inp.Numero,
		Complemento:     inp.Complemento,
		Bairro:          inp.Bairro,
		Cidade:          inp.Cidade,
		UF:              inp.UF,
		CEP:             inp.CEP,
		Bloqueado:       inp.Bloqueado,
		MotivoBloqueio:  inp.MotivoBloqueio,
		Observacoes:     inp.Observacoes,
	}
}

// List GET /api/clientes?search=
func (h *ClientesHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.Store.List(r.Context(), r.URL.Query().Get("search"))
	if err != nil {
		jsonFail(w, http.StatusInternalServerError, "erro ao listar clientes")
		return
	}
	out := make([]clienteJSON, len(list))
	for i, c := range list {
		out[i] = toClienteJSON(c)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":   out,
		"meta":   map[string]any{"total": len(out)},
		"errors": []string{},
	})
}

// Get GET /api/clientes/{id}
func (h *ClientesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonFail(w, http.StatusBadRequest, "id invalido")
		return
	}
	c, err := h.Store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, clientes.ErrNotFound) {
			jsonFail(w, http.StatusNotFound, "cliente nao encontrado")
			return
		}
		jsonFail(w, http.StatusInternalServerError, "erro ao buscar cliente")
		return
	}
	jsonOK(w, toClienteJSON(c))
}

// Create POST /api/clientes
func (h *ClientesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var inp clienteInput
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		jsonFail(w, http.StatusBadRequest, "payload invalido")
		return
	}
	id, err := h.Store.Create(r.Context(), inp.toModel())
	if err != nil {
		switch {
		case errors.Is(err, clientes.ErrDuplicateDoc):
			jsonFail(w, http.StatusConflict, "documento ja cadastrado")
		case errors.Is(err, clientes.ErrInvalidInput):
			jsonFail(w, http.StatusUnprocessableEntity, err.Error())
		default:
			jsonFail(w, http.StatusInternalServerError, "erro ao criar cliente")
		}
		return
	}
	saved, err := h.Store.Get(r.Context(), id)
	if err != nil {
		jsonFail(w, http.StatusInternalServerError, "erro ao buscar cliente criado")
		return
	}
	jsonCreated(w, toClienteJSON(saved))
}

// Update PUT /api/clientes/{id}
func (h *ClientesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonFail(w, http.StatusBadRequest, "id invalido")
		return
	}
	var inp clienteInput
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		jsonFail(w, http.StatusBadRequest, "payload invalido")
		return
	}
	if err := h.Store.Update(r.Context(), id, inp.toModel()); err != nil {
		switch {
		case errors.Is(err, clientes.ErrNotFound):
			jsonFail(w, http.StatusNotFound, "cliente nao encontrado")
		case errors.Is(err, clientes.ErrDuplicateDoc):
			jsonFail(w, http.StatusConflict, "documento ja cadastrado")
		case errors.Is(err, clientes.ErrInvalidInput):
			jsonFail(w, http.StatusUnprocessableEntity, err.Error())
		default:
			jsonFail(w, http.StatusInternalServerError, "erro ao atualizar cliente")
		}
		return
	}
	updated, err := h.Store.Get(r.Context(), id)
	if err != nil {
		jsonFail(w, http.StatusInternalServerError, "erro ao buscar cliente atualizado")
		return
	}
	jsonOK(w, toClienteJSON(updated))
}

// Delete DELETE /api/clientes/{id}
func (h *ClientesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonFail(w, http.StatusBadRequest, "id invalido")
		return
	}
	if err := h.Store.Delete(r.Context(), id); err != nil {
		if errors.Is(err, clientes.ErrNotFound) {
			jsonFail(w, http.StatusNotFound, "cliente nao encontrado")
			return
		}
		jsonFail(w, http.StatusInternalServerError, "erro ao remover cliente")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
