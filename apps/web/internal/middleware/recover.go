package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recover protege os handlers de panicos: imprime stack trace e responde
// 500, evitando que um erro em uma rota derrube o servidor inteiro.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic em %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
				if isAPIRequest(r) {
					writeAPIError(w, http.StatusInternalServerError, "erro interno")
					return
				}
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
