package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrInvalidCredentials e retornado quando login ou senha nao conferem.
var ErrInvalidCredentials = errors.New("credenciais invalidas")

// ErrUserDisabled e retornado quando o usuario existe mas esta inativo.
var ErrUserDisabled = errors.New("usuario desativado")

// User representa um usuario autenticado e suas permissoes resolvidas.
type User struct {
	ID          uuid.UUID
	Login       string
	Email       string
	Nome        string
	Papel       string
	Ativo       bool
	Permissions []string
}

// HasPermission verifica se o usuario possui a permissao informada (ex:
// "agenda.write"). Admin sempre passa.
func (u User) HasPermission(code string) bool {
	if u.Papel == "admin" {
		return true
	}
	for _, p := range u.Permissions {
		if p == code {
			return true
		}
	}
	return false
}

// Store concentra operacoes contra as tabelas usuarios / user_roles /
// role_permissions.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore constroi um novo Store sobre o pool informado.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Authenticate valida org + login + senha e devolve o usuario com permissoes
// resolvidas. Retorna ErrInvalidCredentials em qualquer falha de match,
// para nao vazar se o login existe ou nao.
func (s *Store) Authenticate(ctx context.Context, org, login, password string) (User, error) {
	org = strings.TrimSpace(org)
	login = strings.TrimSpace(login)
	if org == "" || login == "" || password == "" {
		return User{}, ErrInvalidCredentials
	}

	var (
		id    uuid.UUID
		email string
		nome  string
		papel string
		ativo bool
		hash  string
	)
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.nome, u.papel::text, u.ativo, u.senha_hash
		FROM usuarios u
		JOIN tenants t ON t.id = u.tenant_id AND t.slug = $1 AND t.ativo = TRUE
		WHERE (u.login = $2 OR u.email = $2)
		  AND u.deleted_at IS NULL
	`, org, login).Scan(&id, &email, &nome, &papel, &ativo, &hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, fmt.Errorf("buscar usuario: %w", err)
	}

	if !ativo {
		return User{}, ErrUserDisabled
	}
	if err := VerifyPassword(hash, password); err != nil {
		return User{}, ErrInvalidCredentials
	}

	perms, err := s.permissions(ctx, id)
	if err != nil {
		return User{}, err
	}

	_, _ = s.pool.Exec(ctx, `UPDATE usuarios SET ultimo_acesso = NOW() WHERE id = $1`, id)

	return User{
		ID:          id,
		Login:       login,
		Email:       email,
		Nome:        nome,
		Papel:       papel,
		Ativo:       ativo,
		Permissions: perms,
	}, nil
}

// GetByID busca um usuario ja autenticado (via sessao ou JWT) e devolve
// suas permissoes atuais. Usado pelo middleware em cada request para
// re-validar permissoes sem confiar so no token.
func (s *Store) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id, login, email, nome, papel::text, ativo
		FROM usuarios
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&u.ID, &u.Login, &u.Email, &u.Nome, &u.Papel, &u.Ativo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, fmt.Errorf("buscar usuario: %w", err)
	}
	if !u.Ativo {
		return User{}, ErrUserDisabled
	}
	u.Permissions, err = s.permissions(ctx, id)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Store) permissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT p.codigo
		FROM user_roles ur
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE ur.user_id = $1
		ORDER BY p.codigo
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("listar permissoes: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("scan permissao: %w", err)
		}
		out = append(out, code)
	}
	return out, rows.Err()
}

// BootstrapAdmin garante que existe um usuario admin valido no banco. Se
// nao houver nenhum admin, cria com a senha de ADMIN_PASSWORD (ou
// "tsure-admin" como fallback). Idempotente: nao sobrescreve admin
// existente.
func (s *Store) BootstrapAdmin(ctx context.Context) error {
	var count int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM usuarios WHERE papel = 'admin' AND deleted_at IS NULL
	`).Scan(&count); err != nil {
		return fmt.Errorf("contar admins: %w", err)
	}
	if count > 0 {
		return nil
	}

	pw := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
	if pw == "" {
		pw = "tsure-admin"
	}
	hash, err := HashPassword(pw)
	if err != nil {
		return fmt.Errorf("hash admin: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var funcID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO funcionarios (nome, documento, data_admissao, cargo, status)
		VALUES ('Administrador do Sistema', 'ADMIN-BOOTSTRAP', CURRENT_DATE, 'Administrador', 'ativo')
		ON CONFLICT (documento) DO UPDATE SET nome = EXCLUDED.nome
		RETURNING id
	`).Scan(&funcID); err != nil {
		return fmt.Errorf("inserir funcionario admin: %w", err)
	}

	var userID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO usuarios (funcionario_id, login, email, senha_hash, nome, papel, ativo)
		VALUES ($1, 'admin', 'admin@tsure.local', $2, 'Administrador do Sistema', 'admin', TRUE)
		RETURNING id
	`, funcID, hash).Scan(&userID); err != nil {
		return fmt.Errorf("inserir usuario admin: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE codigo = 'admin'
		ON CONFLICT DO NOTHING
	`, userID); err != nil {
		return fmt.Errorf("vincular role admin: %w", err)
	}

	return tx.Commit(ctx)
}
