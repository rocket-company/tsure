// Package middleware reune os middlewares HTTP do tsure: autenticacao web
// (cookie de sessao) e API (JWT bearer), checagem de permissoes RBAC,
// proteccao CSRF (gorilla/csrf) e recuperacao de panico. Os middlewares
// sao desacoplados do router  envolvem qualquer http.Handler.
package middleware

import (
	"context"
	"net/http"

	"tsure/apps/web/internal/auth"
)

type ctxKey int

const (
	userCtxKey  ctxKey = iota
	tokenCtxKey         // token da sessao web, util para revogar no logout
)

// WithUser injeta o usuario autenticado no contexto da request.
func WithUser(ctx context.Context, u auth.User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

// UserFromContext recupera o usuario autenticado, se houver.
func UserFromContext(ctx context.Context) (auth.User, bool) {
	u, ok := ctx.Value(userCtxKey).(auth.User)
	return u, ok
}

// MustUser recupera o usuario do contexto ou despacha 401. Usar apos um
// middleware de autenticacao  se chegou aqui sem usuario, e bug.
func MustUser(r *http.Request) (auth.User, bool) {
	return UserFromContext(r.Context())
}

// WithSessionToken e UseSessionToken sao usados internamente pelo logout
// para recuperar e revogar o token web a partir do contexto.
func WithSessionToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenCtxKey, token)
}

// SessionTokenFromContext devolve o token da sessao web atual, se existir.
func SessionTokenFromContext(ctx context.Context) string {
	if t, ok := ctx.Value(tokenCtxKey).(string); ok {
		return t
	}
	return ""
}
