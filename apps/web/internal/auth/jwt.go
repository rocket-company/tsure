package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Claims e o payload do JWT emitido para a API mobile. Mantemos campos
// canonicos (iss/sub/exp/iat/jti) explicitos para evitar reflexao.
type Claims struct {
	Sub  uuid.UUID `json:"sub"`
	Name string    `json:"name"`
	Role string    `json:"role"`
	Iat  int64     `json:"iat"`
	Exp  int64     `json:"exp"`
	Jti  string    `json:"jti"`
}

// ErrInvalidToken e devolvido quando o token nao pode ser validado.
var ErrInvalidToken = errors.New("invalid token")

// JWTSigner emite e valida JWT HS256 com uma chave compartilhada.
type JWTSigner struct {
	key []byte
	ttl time.Duration
}

// NewJWTSigner cria um signer com chave HMAC e tempo de vida (TTL) dos
// tokens emitidos.
func NewJWTSigner(key []byte, ttl time.Duration) *JWTSigner {
	return &JWTSigner{key: key, ttl: ttl}
}

// Issue gera um JWT HS256 para o usuario informado. Retorna o token
// codificado em base64url e o instante de expiracao.
func (s *JWTSigner) Issue(userID uuid.UUID, name, role string) (string, time.Time, error) {
	now := time.Now().UTC()
	exp := now.Add(s.ttl)
	c := Claims{
		Sub:  userID,
		Name: name,
		Role: role,
		Iat:  now.Unix(),
		Exp:  exp.Unix(),
		Jti:  newJTI(),
	}

	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", time.Time{}, err
	}
	payloadJSON, err := json.Marshal(c)
	if err != nil {
		return "", time.Time{}, err
	}

	enc := base64.RawURLEncoding
	signing := enc.EncodeToString(headerJSON) + "." + enc.EncodeToString(payloadJSON)

	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(signing))
	sig := enc.EncodeToString(mac.Sum(nil))

	return signing + "." + sig, exp, nil
}

// Verify valida assinatura, formato e expiracao de um JWT, retornando os
// claims decodificados. Em qualquer falha devolve ErrInvalidToken envelopado.
func (s *JWTSigner) Verify(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, fmt.Errorf("%w: malformed", ErrInvalidToken)
	}

	enc := base64.RawURLEncoding
	signing := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(signing))
	expected := mac.Sum(nil)

	got, err := enc.DecodeString(parts[2])
	if err != nil {
		return Claims{}, fmt.Errorf("%w: signature decode", ErrInvalidToken)
	}
	if !hmac.Equal(expected, got) {
		return Claims{}, fmt.Errorf("%w: signature mismatch", ErrInvalidToken)
	}

	payload, err := enc.DecodeString(parts[1])
	if err != nil {
		return Claims{}, fmt.Errorf("%w: payload decode", ErrInvalidToken)
	}
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return Claims{}, fmt.Errorf("%w: payload parse", ErrInvalidToken)
	}
	if time.Now().Unix() >= c.Exp {
		return Claims{}, fmt.Errorf("%w: expired", ErrInvalidToken)
	}
	return c, nil
}

// newJTI gera um identificador unico curto. Nao precisa ser criptografico,
// apenas suficiente para distinguir tokens em logs.
func newJTI() string {
	return uuid.New().String()
}
