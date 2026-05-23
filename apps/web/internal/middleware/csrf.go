package middleware

import (
	"net/http"

	"github.com/gorilla/csrf"
)

// CSRFConfig agrupa as opcoes de proteccao Strict CSRF para o BFF web.
// /api/* fica fora porque usa JWT bearer (nao-cookie).
type CSRFConfig struct {
	Key    []byte // 32 bytes
	Secure bool   // true em producao (HTTPS)
}

// NewCSRF devolve um middleware gorilla/csrf configurado em modo Strict:
// SameSite=Strict, HttpOnly, e exige token em todo POST/PUT/DELETE.
// /api/* nao usa cookies; isente esses caminhos antes de aplicar.
func NewCSRF(cfg CSRFConfig) func(http.Handler) http.Handler {
	opts := []csrf.Option{
		csrf.Secure(cfg.Secure),
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.Path("/"),
		csrf.CookieName("tsure_csrf"),
		csrf.FieldName("csrf_token"),
	}
	return csrf.Protect(cfg.Key, opts...)
}

// CSRFTemplateTag injeta o token CSRF nos dados de template. Use no
// handler antes de renderizar o template:
//
//	data["CSRFField"] = middleware.CSRFTemplateTag(r)
func CSRFTemplateTag(r *http.Request) any {
	return csrf.TemplateField(r)
}

// SkipAPICSRF aplica o middleware CSRF apenas a rotas que NAO sao /api/*.
// gorilla/csrf nao tem skip nativo, entao envolvemos manualmente.
func SkipAPICSRF(csrfMW func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		protected := csrfMW(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isAPIRequest(r) {
				next.ServeHTTP(w, r)
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}
