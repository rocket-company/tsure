// Package servicos implementa o catalogo de servicos de locacao  CRUD
// sobre servicos_locacao + lookup de classificacoes_servico. Cada servico
// e um item ofertavel que entra como linha em uma agenda (OS).
package servicos

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound     = errors.New("servico nao encontrado")
	ErrDuplicate    = errors.New("codigo de servico ja cadastrado")
	ErrInvalidInput = errors.New("dados invalidos")
)

// Servico representa uma linha de servicos_locacao com a classificacao
// resolvida via JOIN.
type Servico struct {
	ID                     uuid.UUID
	Codigo                 string
	Descricao              string
	ClassificacaoID        *uuid.UUID
	ClassificacaoCodigo    string
	ClassificacaoDescricao string
	UnidadePadrao          string
	ValorReferencia        float64
	Ativo                  bool
}

// Classificacao e uma categoria macro (Palco, Tenda, Sonorizacao, ...).
type Classificacao struct {
	ID        uuid.UUID
	Codigo    string
	Descricao string
}

// Store concentra acessos a servicos_locacao e classificacoes_servico.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore cria um Store sobre o pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// List devolve servicos filtrados por codigo OU descricao (case-insensitive).
func (s *Store) List(ctx context.Context, query string) ([]Servico, error) {
	q := strings.TrimSpace(query)
	args := []any{}
	where := ""
	if q != "" {
		args = append(args, "%"+strings.ToLower(q)+"%")
		where = "WHERE (lower(sl.descricao) LIKE $1 OR lower(sl.codigo) LIKE $1)"
	}
	sql := `
		SELECT sl.id, sl.codigo, sl.descricao, sl.classificacao_id,
		       COALESCE(c.codigo, ''), COALESCE(c.descricao, ''),
		       sl.unidade_padrao, sl.valor_referencia, sl.ativo
		FROM servicos_locacao sl
		LEFT JOIN classificacoes_servico c ON c.id = sl.classificacao_id
		` + where + `
		ORDER BY
		  CASE WHEN sl.codigo ~ '^[0-9]+$' THEN LPAD(sl.codigo, 10, '0') ELSE sl.codigo END
		LIMIT 500
	`
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("listar servicos: %w", err)
	}
	defer rows.Close()

	out := make([]Servico, 0, 32)
	for rows.Next() {
		var sv Servico
		if err := rows.Scan(
			&sv.ID, &sv.Codigo, &sv.Descricao, &sv.ClassificacaoID,
			&sv.ClassificacaoCodigo, &sv.ClassificacaoDescricao,
			&sv.UnidadePadrao, &sv.ValorReferencia, &sv.Ativo,
		); err != nil {
			return nil, fmt.Errorf("scan servico: %w", err)
		}
		out = append(out, sv)
	}
	return out, rows.Err()
}

// Get busca um servico por id.
func (s *Store) Get(ctx context.Context, id uuid.UUID) (Servico, error) {
	var sv Servico
	err := s.pool.QueryRow(ctx, `
		SELECT sl.id, sl.codigo, sl.descricao, sl.classificacao_id,
		       COALESCE(c.codigo, ''), COALESCE(c.descricao, ''),
		       sl.unidade_padrao, sl.valor_referencia, sl.ativo
		FROM servicos_locacao sl
		LEFT JOIN classificacoes_servico c ON c.id = sl.classificacao_id
		WHERE sl.id = $1
	`, id).Scan(
		&sv.ID, &sv.Codigo, &sv.Descricao, &sv.ClassificacaoID,
		&sv.ClassificacaoCodigo, &sv.ClassificacaoDescricao,
		&sv.UnidadePadrao, &sv.ValorReferencia, &sv.Ativo,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Servico{}, ErrNotFound
		}
		return Servico{}, fmt.Errorf("buscar servico: %w", err)
	}
	return sv, nil
}

// Create insere um servico  se codigo vazio, gera proximo sequencial.
func (s *Store) Create(ctx context.Context, sv Servico) (uuid.UUID, error) {
	if err := sv.Validate(); err != nil {
		return uuid.Nil, err
	}
	if strings.TrimSpace(sv.Codigo) == "" {
		c, err := s.nextCodigo(ctx)
		if err != nil {
			return uuid.Nil, err
		}
		sv.Codigo = c
	}
	if strings.TrimSpace(sv.UnidadePadrao) == "" {
		sv.UnidadePadrao = "DIARIA"
	}
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO servicos_locacao
		    (codigo, descricao, classificacao_id, unidade_padrao, valor_referencia, ativo)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, sv.Codigo, sv.Descricao, sv.ClassificacaoID, sv.UnidadePadrao, sv.ValorReferencia, sv.Ativo).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return uuid.Nil, ErrDuplicate
		}
		return uuid.Nil, fmt.Errorf("inserir servico: %w", err)
	}
	return id, nil
}

// Update sobrescreve um servico existente.
func (s *Store) Update(ctx context.Context, id uuid.UUID, sv Servico) error {
	if err := sv.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(sv.UnidadePadrao) == "" {
		sv.UnidadePadrao = "DIARIA"
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE servicos_locacao SET
		    codigo = $2,
		    descricao = $3,
		    classificacao_id = $4,
		    unidade_padrao = $5,
		    valor_referencia = $6,
		    ativo = $7
		WHERE id = $1
	`, id, sv.Codigo, sv.Descricao, sv.ClassificacaoID, sv.UnidadePadrao, sv.ValorReferencia, sv.Ativo)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicate
		}
		return fmt.Errorf("atualizar servico: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete e soft-delete: marca ativo=false. Servicos referenciados em
// agenda_itens nao podem ser removidos fisicamente.
func (s *Store) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `UPDATE servicos_locacao SET ativo = FALSE WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("desativar servico: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListClassificacoes retorna as classificacoes ativas para popular o
// dropdown do formulario.
func (s *Store) ListClassificacoes(ctx context.Context) ([]Classificacao, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, codigo, descricao FROM classificacoes_servico
		WHERE ativo = TRUE
		ORDER BY ordem, descricao
	`)
	if err != nil {
		return nil, fmt.Errorf("listar classificacoes: %w", err)
	}
	defer rows.Close()
	out := make([]Classificacao, 0, 16)
	for rows.Next() {
		var c Classificacao
		if err := rows.Scan(&c.ID, &c.Codigo, &c.Descricao); err != nil {
			return nil, fmt.Errorf("scan classificacao: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// nextCodigo devolve o proximo codigo numerico sequencial disponivel.
func (s *Store) nextCodigo(ctx context.Context) (string, error) {
	var max int
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(codigo::int), 0)
		FROM servicos_locacao
		WHERE codigo ~ '^[0-9]+$'
	`).Scan(&max)
	if err != nil {
		return "", fmt.Errorf("next codigo: %w", err)
	}
	return strconv.Itoa(max + 1), nil
}

// Validate confere regras minimas.
func (sv Servico) Validate() error {
	if strings.TrimSpace(sv.Descricao) == "" {
		return fmt.Errorf("%w: descricao obrigatoria", ErrInvalidInput)
	}
	return nil
}

// ValorFormatado devolve o valor em BRL para exibicao.
func (sv Servico) ValorFormatado() string {
	return fmt.Sprintf("R$ %.2f", sv.ValorReferencia)
}

func isUniqueViolation(err error) bool {
	type pgErr interface{ SQLState() string }
	if pe, ok := err.(pgErr); ok {
		return pe.SQLState() == "23505"
	}
	return false
}
