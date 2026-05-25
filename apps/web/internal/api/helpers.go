// Package api implementa os handlers da API JSON consumida pelo app mobile.
// Todas as rotas ficam sob /api/* e usam JWT bearer para autenticacao.
package api

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func jsonOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, map[string]any{
		"data":    data,
		"message": "ok",
		"errors":  []string{},
	})
}

func jsonCreated(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, map[string]any{
		"data":    data,
		"message": "ok",
		"errors":  []string{},
	})
}

func jsonFail(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{
		"data":    nil,
		"message": msg,
		"errors":  []string{msg},
	})
}
