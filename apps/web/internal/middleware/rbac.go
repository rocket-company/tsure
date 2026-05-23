package middleware

import "net/http"

// RequirePermission devolve um middleware que checa se o usuario do
// contexto possui a permissao informada (ex: "agenda.write"). Use apos
// WebAuth.Required ou APIAuth.Required.
func RequirePermission(code string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok {
				writeAccessDenied(w, r, "nao autenticado", http.StatusUnauthorized)
				return
			}
			if !u.HasPermission(code) {
				writeAccessDenied(w, r, "permissao negada: "+code, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission passa adiante quando o usuario tem PELO MENOS UMA
// das permissoes listadas. Util para telas de leitura que aceitam varios
// niveis de acesso.
func RequireAnyPermission(codes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok {
				writeAccessDenied(w, r, "nao autenticado", http.StatusUnauthorized)
				return
			}
			for _, c := range codes {
				if u.HasPermission(c) {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeAccessDenied(w, r, "permissao negada", http.StatusForbidden)
		})
	}
}

func writeAccessDenied(w http.ResponseWriter, r *http.Request, msg string, status int) {
	if isAPIRequest(r) {
		writeAPIError(w, status, msg)
		return
	}
	http.Error(w, msg, status)
}

func isAPIRequest(r *http.Request) bool {
	if r.URL == nil {
		return false
	}
	return len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/"
}
