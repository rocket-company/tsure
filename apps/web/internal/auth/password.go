// Package auth concentra emissao e verificacao de credenciais para o ERP:
// hash de senha (bcrypt), sessoes opacas para o BFF web e JWT para a API
// consumida pelo mobile. Nao acopla a transporte (HTTP): handlers e
// middlewares vivem em internal/middleware.
package auth

import "golang.org/x/crypto/bcrypt"

// HashCost e o custo bcrypt usado em todas as senhas. 12 e um bom meio
// termo entre seguranca e latencia em hardware moderno.
const HashCost = 12

// HashPassword gera um hash bcrypt para a senha em texto claro.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), HashCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// VerifyPassword retorna nil quando a senha bate com o hash, ou um erro
// (geralmente bcrypt.ErrMismatchedHashAndPassword) caso contrario.
func VerifyPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
