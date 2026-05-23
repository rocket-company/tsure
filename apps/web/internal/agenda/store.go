// Package agenda implementa o nucleo operacional do ERP: agendamento de
// servicos (que tambem e a Ordem de Servico, identificada pelo campo
// numero auto-incrementado). Persiste em agenda + agenda_locais +
// agenda_local_contatos.
package agenda

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// refCacheTTL define por quanto tempo as listas de clientes/funcionarios
// usadas em dropdowns ficam cacheadas em memoria. Mudancas pos-TTL
// aparecem no proximo render.
const refCacheTTL = 60 * time.Second

var (
	ErrNotFound     = errors.New("agendamento nao encontrado")
	ErrInvalidInput = errors.New("dados invalidos")
)

// Status validos do agendamento (espelham agenda_status).
const (
	StatusOrcamento         = "orcamento"
	StatusEmAnalise         = "em_analise"
	StatusAprovado          = "aprovado"
	StatusAgendado          = "agendado"
	StatusEmExecucao        = "em_execucao"
	StatusAguardandoRetorno = "aguardando_retorno"
	StatusFinalizado        = "finalizado"
	StatusCancelado         = "cancelado"
)

// Tipos de evento espelham agenda_tipo_evento.
const (
	TipoParticular  = "particular"
	TipoLicitacao   = "licitacao"
	TipoCortesia    = "cortesia"
	TipoRecorrente  = "recorrente"
)

// Agendamento e a visao consolidada de uma OS para o formulario  inclui
// joins de cliente (nome/endereco), funcionario que negociou ("quem
// contratou"), local principal e os 2 primeiros contatos cadastrados.
type Agendamento struct {
	ID                  uuid.UUID
	Numero              int64
	ClienteID           uuid.UUID
	ClienteNome         string
	ClienteDocumento    string
	QuemContratouID     *uuid.UUID
	QuemContratouNome   string
	Status              string
	TipoEvento          string
	DescricaoEvento     string
	DataEvento          *time.Time
	HoraEvento          string
	DataInstalacao      *time.Time
	HoraInstalacao      string
	DataRetornoPrevista *time.Time
	FormaPagamento      string
	ValorTotal          float64
	NumeroAprovacao     string
	DataAprovacao       *time.Time
	UsuarioAprovadorID  *uuid.UUID
	QuemAprovou         string
	Observacoes         string
	Finalizado          bool

	// agenda_locais (principal)
	LocalLogradouro  string
	LocalNumero      string
	LocalComplemento string
	LocalBairro      string
	LocalCidade      string
	LocalUF          string

	// agenda_local_contatos (ate 2)
	Responsavel1Nome     string
	Responsavel1Telefone string
	Responsavel2Nome     string
	Responsavel2Telefone string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ClienteRef e a versao enxuta para popular dropdowns + auto-fill de endereco.
type ClienteRef struct {
	ID         uuid.UUID
	Nome       string
	Documento  string
	Logradouro string
	Numero     string
	Complemento string
	Bairro     string
	Cidade     string
	UF         string
}

// FuncionarioRef e a versao enxuta para o dropdown de "Quem contratou".
type FuncionarioRef struct {
	ID    uuid.UUID
	Nome  string
	Cargo string
}

// Store concentra operacoes sobre agenda/agenda_locais/agenda_local_contatos.
// Mantem cache em memoria das listas de referencia (clientes/funcionarios)
// para reduzir round-trips na WAN  cada uma com seu mutex.
type Store struct {
	pool *pgxpool.Pool

	clientesMu     sync.Mutex
	clientesCache  []ClienteRef
	clientesExp    time.Time

	funcionariosMu     sync.Mutex
	funcionariosCache  []FuncionarioRef
	funcionariosExp    time.Time
}

// NewStore cria um Store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// InvalidateRefs limpa os caches de listas auxiliares. Chamar quando um
// cliente ou funcionario for criado/editado/removido.
func (s *Store) InvalidateRefs() {
	s.clientesMu.Lock()
	s.clientesCache, s.clientesExp = nil, time.Time{}
	s.clientesMu.Unlock()

	s.funcionariosMu.Lock()
	s.funcionariosCache, s.funcionariosExp = nil, time.Time{}
	s.funcionariosMu.Unlock()
}

// List devolve os agendamentos mais recentes (limite 200), filtrados por
// nome do cliente ou numero da OS.
func (s *Store) List(ctx context.Context, query string) ([]Agendamento, error) {
	q := strings.TrimSpace(query)
	args := []any{}
	where := "WHERE a.deleted_at IS NULL"
	if q != "" {
		args = append(args, "%"+strings.ToLower(q)+"%")
		where += " AND (lower(c.nome_razao_social) LIKE $1 OR CAST(a.numero AS text) LIKE $1)"
	}
	sql := `
		SELECT a.id, a.numero, a.cliente_id, c.nome_razao_social, c.documento,
		       a.quem_contratou_id, COALESCE(f.nome, ''),
		       a.status::text, a.tipo_evento::text,
		       COALESCE(a.descricao_evento, ''),
		       a.data_evento, a.data_instalacao,
		       COALESCE(a.forma_pagamento::text, ''),
		       a.valor_total,
		       a.finalizado,
		       a.created_at
		FROM agenda a
		JOIN clientes c ON c.id = a.cliente_id
		LEFT JOIN funcionarios f ON f.id = a.quem_contratou_id
		` + where + `
		ORDER BY a.numero DESC
		LIMIT 200
	`
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("listar agenda: %w", err)
	}
	defer rows.Close()

	out := make([]Agendamento, 0, 32)
	for rows.Next() {
		var a Agendamento
		if err := rows.Scan(
			&a.ID, &a.Numero, &a.ClienteID, &a.ClienteNome, &a.ClienteDocumento,
			&a.QuemContratouID, &a.QuemContratouNome,
			&a.Status, &a.TipoEvento,
			&a.DescricaoEvento,
			&a.DataEvento, &a.DataInstalacao,
			&a.FormaPagamento,
			&a.ValorTotal,
			&a.Finalizado,
			&a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan agenda: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Get devolve um agendamento completo com local principal e dois primeiros
// contatos.
func (s *Store) Get(ctx context.Context, id uuid.UUID) (Agendamento, error) {
	var a Agendamento
	var (
		horaEvento     *string
		horaInstalacao *string
		formaPag       *string
	)
	err := s.pool.QueryRow(ctx, `
		SELECT a.id, a.numero, a.cliente_id, c.nome_razao_social, c.documento,
		       a.quem_contratou_id, COALESCE(f.nome, ''),
		       a.status::text, a.tipo_evento::text,
		       COALESCE(a.descricao_evento, ''),
		       a.data_evento, to_char(a.hora_evento, 'HH24:MI'),
		       a.data_instalacao, to_char(a.hora_instalacao, 'HH24:MI'),
		       a.data_retorno_prevista,
		       a.forma_pagamento::text,
		       a.valor_total,
		       COALESCE(a.numero_aprovacao, ''),
		       a.data_aprovacao,
		       a.usuario_aprovador_id,
		       COALESCE(u.nome, ''),
		       COALESCE(a.observacoes, ''),
		       a.finalizado,
		       a.created_at, a.updated_at
		FROM agenda a
		JOIN clientes c ON c.id = a.cliente_id
		LEFT JOIN usuarios u ON u.id = a.usuario_aprovador_id
		LEFT JOIN funcionarios f ON f.id = a.quem_contratou_id
		WHERE a.id = $1 AND a.deleted_at IS NULL
	`, id).Scan(
		&a.ID, &a.Numero, &a.ClienteID, &a.ClienteNome, &a.ClienteDocumento,
		&a.QuemContratouID, &a.QuemContratouNome,
		&a.Status, &a.TipoEvento,
		&a.DescricaoEvento,
		&a.DataEvento, &horaEvento,
		&a.DataInstalacao, &horaInstalacao,
		&a.DataRetornoPrevista,
		&formaPag,
		&a.ValorTotal,
		&a.NumeroAprovacao,
		&a.DataAprovacao,
		&a.UsuarioAprovadorID,
		&a.QuemAprovou,
		&a.Observacoes,
		&a.Finalizado,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Agendamento{}, ErrNotFound
		}
		return Agendamento{}, fmt.Errorf("buscar agenda: %w", err)
	}
	if horaEvento != nil {
		a.HoraEvento = *horaEvento
	}
	if horaInstalacao != nil {
		a.HoraInstalacao = *horaInstalacao
	}
	if formaPag != nil {
		a.FormaPagamento = *formaPag
	}

	// Local principal
	var localID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT id, COALESCE(logradouro,''), COALESCE(numero,''), COALESCE(complemento,''),
		       COALESCE(bairro,''), COALESCE(cidade,''), COALESCE(uf,'')
		FROM agenda_locais
		WHERE agenda_id = $1 AND principal = TRUE
		LIMIT 1
	`, a.ID).Scan(
		&localID,
		&a.LocalLogradouro, &a.LocalNumero, &a.LocalComplemento,
		&a.LocalBairro, &a.LocalCidade, &a.LocalUF,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Agendamento{}, fmt.Errorf("buscar local: %w", err)
	}

	if err == nil {
		// Ate 2 contatos do local principal
		rows, err := s.pool.Query(ctx, `
			SELECT COALESCE(nome,''), COALESCE(telefone_principal,'')
			FROM agenda_local_contatos
			WHERE agenda_local_id = $1
			ORDER BY principal DESC, created_at ASC
			LIMIT 2
		`, localID)
		if err != nil {
			return Agendamento{}, fmt.Errorf("buscar contatos: %w", err)
		}
		defer rows.Close()
		idx := 0
		for rows.Next() {
			var nome, tel string
			if err := rows.Scan(&nome, &tel); err != nil {
				return Agendamento{}, fmt.Errorf("scan contato: %w", err)
			}
			if idx == 0 {
				a.Responsavel1Nome, a.Responsavel1Telefone = nome, tel
			} else {
				a.Responsavel2Nome, a.Responsavel2Telefone = nome, tel
			}
			idx++
		}
	}
	return a, nil
}

// Save persiste um agendamento  cria se ID == uuid.Nil, atualiza caso
// contrario. Toda a operacao roda em uma transacao para garantir
// consistencia entre agenda, agenda_locais e agenda_local_contatos.
func (s *Store) Save(ctx context.Context, a Agendamento, usuarioRegistro *uuid.UUID) (uuid.UUID, int64, error) {
	if err := a.Validate(); err != nil {
		return uuid.Nil, 0, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		agendaID uuid.UUID
		numero   int64
	)

	if a.ID == uuid.Nil {
		// CREATE
		err = tx.QueryRow(ctx, `
			INSERT INTO agenda (
				cliente_id, usuario_registro_id, quem_contratou_id,
				status, tipo_evento,
				descricao_evento,
				data_evento, hora_evento,
				data_instalacao, hora_instalacao,
				data_retorno_prevista,
				forma_pagamento,
				valor_total,
				numero_aprovacao, data_aprovacao,
				observacoes,
				finalizado
			) VALUES (
				$1, $2, $3,
				$4::agenda_status, $5::agenda_tipo_evento,
				$6,
				$7, NULLIF($8, '')::time,
				$9, NULLIF($10, '')::time,
				$11,
				NULLIF($12, '')::forma_pagamento,
				$13,
				$14, $15,
				$16,
				$17
			)
			RETURNING id, numero
		`,
			a.ClienteID, usuarioRegistro, a.QuemContratouID,
			nonEmptyOr(a.Status, StatusOrcamento), nonEmptyOr(a.TipoEvento, TipoParticular),
			a.DescricaoEvento,
			a.DataEvento, a.HoraEvento,
			a.DataInstalacao, a.HoraInstalacao,
			a.DataRetornoPrevista,
			a.FormaPagamento,
			a.ValorTotal,
			a.NumeroAprovacao, a.DataAprovacao,
			a.Observacoes,
			a.Finalizado,
		).Scan(&agendaID, &numero)
		if err != nil {
			return uuid.Nil, 0, fmt.Errorf("inserir agenda: %w", err)
		}
	} else {
		// UPDATE
		agendaID = a.ID
		numero = a.Numero
		_, err = tx.Exec(ctx, `
			UPDATE agenda SET
				cliente_id = $2,
				quem_contratou_id = $3,
				status = $4::agenda_status,
				tipo_evento = $5::agenda_tipo_evento,
				descricao_evento = $6,
				data_evento = $7, hora_evento = NULLIF($8, '')::time,
				data_instalacao = $9, hora_instalacao = NULLIF($10, '')::time,
				data_retorno_prevista = $11,
				forma_pagamento = NULLIF($12, '')::forma_pagamento,
				valor_total = $13,
				numero_aprovacao = $14,
				data_aprovacao = $15,
				observacoes = $16,
				finalizado = $17
			WHERE id = $1 AND deleted_at IS NULL
		`,
			agendaID,
			a.ClienteID, a.QuemContratouID,
			nonEmptyOr(a.Status, StatusOrcamento), nonEmptyOr(a.TipoEvento, TipoParticular),
			a.DescricaoEvento,
			a.DataEvento, a.HoraEvento,
			a.DataInstalacao, a.HoraInstalacao,
			a.DataRetornoPrevista,
			a.FormaPagamento,
			a.ValorTotal,
			a.NumeroAprovacao, a.DataAprovacao,
			a.Observacoes,
			a.Finalizado,
		)
		if err != nil {
			return uuid.Nil, 0, fmt.Errorf("atualizar agenda: %w", err)
		}
	}

	// Local principal (estrategia: drop+recreate para simplificar reedicoes)
	if _, err := tx.Exec(ctx, `DELETE FROM agenda_locais WHERE agenda_id = $1`, agendaID); err != nil {
		return uuid.Nil, 0, fmt.Errorf("limpar locais: %w", err)
	}
	var localID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO agenda_locais (
			agenda_id, tipo, logradouro, numero, complemento, bairro, cidade, uf, principal
		) VALUES (
			$1, 'principal', $2, $3, $4, $5, $6, NULLIF($7, ''), TRUE
		)
		RETURNING id
	`, agendaID,
		a.LocalLogradouro, a.LocalNumero, a.LocalComplemento,
		a.LocalBairro, a.LocalCidade, a.LocalUF,
	).Scan(&localID); err != nil {
		return uuid.Nil, 0, fmt.Errorf("inserir local: %w", err)
	}

	// Responsaveis (somente nao-vazios)
	if strings.TrimSpace(a.Responsavel1Nome) != "" || strings.TrimSpace(a.Responsavel1Telefone) != "" {
		if _, err := tx.Exec(ctx, `
			INSERT INTO agenda_local_contatos (agenda_local_id, nome, telefone_principal, principal)
			VALUES ($1, $2, $3, TRUE)
		`, localID, a.Responsavel1Nome, a.Responsavel1Telefone); err != nil {
			return uuid.Nil, 0, fmt.Errorf("inserir contato 1: %w", err)
		}
	}
	if strings.TrimSpace(a.Responsavel2Nome) != "" || strings.TrimSpace(a.Responsavel2Telefone) != "" {
		if _, err := tx.Exec(ctx, `
			INSERT INTO agenda_local_contatos (agenda_local_id, nome, telefone_principal, principal)
			VALUES ($1, $2, $3, FALSE)
		`, localID, a.Responsavel2Nome, a.Responsavel2Telefone); err != nil {
			return uuid.Nil, 0, fmt.Errorf("inserir contato 2: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, 0, fmt.Errorf("commit: %w", err)
	}
	return agendaID, numero, nil
}

// Delete e soft-delete da agenda.
func (s *Store) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `UPDATE agenda SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("remover agenda: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListClientesRef devolve a lista enxuta de clientes (ate 500) para o
// dropdown de selecao + auto-fill do endereco. Cacheada por refCacheTTL.
func (s *Store) ListClientesRef(ctx context.Context) ([]ClienteRef, error) {
	s.clientesMu.Lock()
	defer s.clientesMu.Unlock()
	if s.clientesCache != nil && time.Now().Before(s.clientesExp) {
		return s.clientesCache, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, nome_razao_social, documento,
		       COALESCE(logradouro,''), COALESCE(numero,''), COALESCE(complemento,''),
		       COALESCE(bairro,''), COALESCE(cidade,''), COALESCE(uf,'')
		FROM clientes
		WHERE deleted_at IS NULL AND bloqueado = FALSE
		ORDER BY nome_razao_social
		LIMIT 500
	`)
	if err != nil {
		return nil, fmt.Errorf("listar clientes ref: %w", err)
	}
	defer rows.Close()
	out := make([]ClienteRef, 0, 64)
	for rows.Next() {
		var c ClienteRef
		if err := rows.Scan(
			&c.ID, &c.Nome, &c.Documento,
			&c.Logradouro, &c.Numero, &c.Complemento,
			&c.Bairro, &c.Cidade, &c.UF,
		); err != nil {
			return nil, fmt.Errorf("scan cliente ref: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	s.clientesCache = out
	s.clientesExp = time.Now().Add(refCacheTTL)
	return out, nil
}

// ListFuncionariosRef devolve a lista enxuta de funcionarios ativos para
// popular o dropdown de "Quem contratou evento". Cacheada por refCacheTTL.
func (s *Store) ListFuncionariosRef(ctx context.Context) ([]FuncionarioRef, error) {
	s.funcionariosMu.Lock()
	defer s.funcionariosMu.Unlock()
	if s.funcionariosCache != nil && time.Now().Before(s.funcionariosExp) {
		return s.funcionariosCache, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, nome, COALESCE(cargo, '')
		FROM funcionarios
		WHERE deleted_at IS NULL AND status = 'ativo'
		ORDER BY nome
		LIMIT 500
	`)
	if err != nil {
		return nil, fmt.Errorf("listar funcionarios ref: %w", err)
	}
	defer rows.Close()
	out := make([]FuncionarioRef, 0, 32)
	for rows.Next() {
		var f FuncionarioRef
		if err := rows.Scan(&f.ID, &f.Nome, &f.Cargo); err != nil {
			return nil, fmt.Errorf("scan funcionario ref: %w", err)
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	s.funcionariosCache = out
	s.funcionariosExp = time.Now().Add(refCacheTTL)
	return out, nil
}

// GetClienteRef devolve um cliente pelo id (para pre-carga do form com
// query param ?cliente=). Aceita id zero (retorna ClienteRef{} sem erro).
func (s *Store) GetClienteRef(ctx context.Context, id uuid.UUID) (ClienteRef, error) {
	if id == uuid.Nil {
		return ClienteRef{}, nil
	}
	var c ClienteRef
	err := s.pool.QueryRow(ctx, `
		SELECT id, nome_razao_social, documento,
		       COALESCE(logradouro,''), COALESCE(numero,''), COALESCE(complemento,''),
		       COALESCE(bairro,''), COALESCE(cidade,''), COALESCE(uf,'')
		FROM clientes
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&c.ID, &c.Nome, &c.Documento,
		&c.Logradouro, &c.Numero, &c.Complemento,
		&c.Bairro, &c.Cidade, &c.UF,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ClienteRef{}, nil
		}
		return ClienteRef{}, fmt.Errorf("buscar cliente ref: %w", err)
	}
	return c, nil
}

// EnderecoCompleto monta uma linha de exibicao do endereco do cliente.
func (c ClienteRef) EnderecoCompleto() string {
	parts := make([]string, 0, 6)
	if c.Logradouro != "" {
		v := c.Logradouro
		if c.Numero != "" {
			v += ", " + c.Numero
		}
		parts = append(parts, v)
	}
	if c.Complemento != "" {
		parts = append(parts, c.Complemento)
	}
	if c.Bairro != "" {
		parts = append(parts, c.Bairro)
	}
	if c.Cidade != "" {
		ufStr := c.Cidade
		if c.UF != "" {
			ufStr += "/" + c.UF
		}
		parts = append(parts, ufStr)
	}
	return strings.Join(parts, "  ")
}

// Validate confere regras minimas de negocio antes de persistir.
func (a Agendamento) Validate() error {
	if a.ClienteID == uuid.Nil {
		return fmt.Errorf("%w: cliente obrigatorio", ErrInvalidInput)
	}
	if a.DataEvento != nil && a.DataInstalacao != nil {
		if a.DataEvento.Before(*a.DataInstalacao) {
			return fmt.Errorf("%w: data do evento nao pode ser anterior a data de instalacao", ErrInvalidInput)
		}
	}
	return nil
}

// StatusLabel devolve o rotulo humano do status.
func (a Agendamento) StatusLabel() string {
	switch a.Status {
	case StatusOrcamento:
		return "Orcamento"
	case StatusEmAnalise:
		return "Em analise"
	case StatusAprovado:
		return "Aprovado"
	case StatusAgendado:
		return "Agendado"
	case StatusEmExecucao:
		return "Em execucao"
	case StatusAguardandoRetorno:
		return "Aguardando retorno"
	case StatusFinalizado:
		return "Finalizado"
	case StatusCancelado:
		return "Cancelado"
	default:
		return a.Status
	}
}

// NumeroFormatado devolve a OS com prefixo OS- e zero-padding.
func (a Agendamento) NumeroFormatado() string {
	if a.Numero == 0 {
		return "-"
	}
	return fmt.Sprintf("OS-%06d", a.Numero)
}

func nonEmptyOr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
