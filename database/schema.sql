-- =============================================================================
-- tsure  ERP de Locacoes e Leasing de Projetos — Multi-Tenant
-- Schema canonico PostgreSQL, derivado de docs/ERD.md.
-- Convencoes: UUIDv7 PK, snake_case, timestamptz, soft-delete (deleted_at),
-- ENUMs nativos, tenant_id em todas as entidades de negocio.
--
-- Login estilo AWS IAM:  slug_do_tenant/login_usuario  (ex: radelgo/admin)
-- Papeis do sistema (tenant_id IS NULL) estao em roles.sistema = TRUE.
-- =============================================================================

SET client_min_messages = WARNING;

-- -----------------------------------------------------------------------------
-- Extensoes
-- -----------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

-- UUIDv7 (Postgres 16 nao tem built-in: implementacao baseada em RFC 9562).
CREATE OR REPLACE FUNCTION uuidv7() RETURNS uuid AS $$
DECLARE
    ts_ms      bytea;
    rand_bytes bytea;
    uuid_bytes bytea;
BEGIN
    ts_ms      := substring(int8send((extract(epoch FROM clock_timestamp()) * 1000)::bigint) FROM 3);
    rand_bytes := gen_random_bytes(10);
    uuid_bytes := ts_ms || rand_bytes;
    uuid_bytes := set_byte(uuid_bytes, 6, ((b'01110000'::int) | (get_byte(uuid_bytes, 6) & 15)));
    uuid_bytes := set_byte(uuid_bytes, 8, ((b'10000000'::int) | (get_byte(uuid_bytes, 8) & 63)));
    RETURN encode(uuid_bytes, 'hex')::uuid;
END;
$$ LANGUAGE plpgsql VOLATILE;

-- Trigger reusavel para updated_at
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- -----------------------------------------------------------------------------
-- ENUMs
-- -----------------------------------------------------------------------------
DO $$ BEGIN
    CREATE TYPE cliente_tipo AS ENUM ('pessoa_fisica', 'pessoa_juridica');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE funcionario_status AS ENUM ('ativo', 'desligado', 'afastado');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE usuario_papel AS ENUM ('admin', 'comercial', 'operacao', 'financeiro', 'fiscal', 'campo');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE agenda_status AS ENUM (
        'orcamento', 'em_analise', 'aprovado', 'agendado',
        'em_execucao', 'aguardando_retorno', 'finalizado', 'cancelado'
    );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE agenda_tipo_evento AS ENUM ('particular', 'licitacao', 'cortesia', 'recorrente');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE agenda_tipo_retorno AS ENUM ('mesma_equipe', 'outra_equipe');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE agenda_local_tipo AS ENUM ('principal', 'montagem', 'apoio', 'hospedagem', 'estacionamento');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE agenda_equipe_papel AS ENUM ('instalacao', 'retorno');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE veiculo_status AS ENUM ('disponivel', 'reservado', 'em_rota', 'em_manutencao', 'indisponivel');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE equipamento_status AS ENUM ('disponivel', 'em_uso', 'em_manutencao', 'baixado');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE movimentacao_tipo AS ENUM ('saida', 'retorno', 'avaria', 'perda', 'baixa', 'manutencao');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE forma_pagamento AS ENUM ('dinheiro', 'cheque', 'transferencia', 'pix', 'boleto', 'cartao');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE tipo_documento_fiscal AS ENUM ('recibo', 'nota_fiscal', 'cupom');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE conta_status AS ENUM ('previsto', 'faturado', 'em_aberto', 'parcial', 'pago', 'renegociado', 'cancelado');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- =============================================================================
-- TENANTS — cada empresa/white-label e um tenant isolado.
-- Login estilo IAM:  slug/usuario  (ex: radelgo/admin)
-- =============================================================================
CREATE TABLE IF NOT EXISTS tenants (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    slug        varchar(60) NOT NULL UNIQUE,
    nome        varchar(200) NOT NULL,
    dominio     varchar(200),
    plano       varchar(40) NOT NULL DEFAULT 'standard',
    ativo       boolean NOT NULL DEFAULT TRUE,
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW(),
    deleted_at  timestamptz
);

DROP TRIGGER IF EXISTS tenants_set_updated_at ON tenants;
CREATE TRIGGER tenants_set_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Sequencias numericas por tenant e entidade (substitui sequences globais).
CREATE TABLE IF NOT EXISTS tenant_sequences (
    tenant_id   uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entidade    varchar(40) NOT NULL,
    ultimo_seq  bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, entidade)
);

CREATE OR REPLACE FUNCTION next_tenant_seq(p_tenant_id uuid, p_entidade varchar)
RETURNS bigint AS $$
DECLARE v_next bigint;
BEGIN
    INSERT INTO tenant_sequences (tenant_id, entidade, ultimo_seq)
    VALUES (p_tenant_id, p_entidade, 1)
    ON CONFLICT (tenant_id, entidade)
    DO UPDATE SET ultimo_seq = tenant_sequences.ultimo_seq + 1
    RETURNING ultimo_seq INTO v_next;
    RETURN v_next;
END;
$$ LANGUAGE plpgsql;

-- Trigger: gera numero sequencial por tenant em agenda.
CREATE OR REPLACE FUNCTION agenda_before_insert() RETURNS trigger AS $$
BEGIN
    IF NEW.numero IS NULL OR NEW.numero = 0 THEN
        NEW.numero := next_tenant_seq(NEW.tenant_id, 'agenda');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: gera numero_titulo por tenant em contas_receber.
CREATE OR REPLACE FUNCTION contas_receber_before_insert() RETURNS trigger AS $$
BEGIN
    IF NEW.numero_titulo IS NULL OR NEW.numero_titulo = '' THEN
        NEW.numero_titulo := 'CR-' || lpad(next_tenant_seq(NEW.tenant_id, 'contas_receber')::text, 8, '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- RBAC  — roles com tenant_id nullable:
--   NULL  = role de sistema (compartilhado, gerenciado pelo produto)
--   uuid  = role customizado do tenant
-- =============================================================================
CREATE TABLE IF NOT EXISTS roles (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id   uuid REFERENCES tenants(id) ON DELETE CASCADE,
    codigo      varchar(40) NOT NULL,
    descricao   varchar(120) NOT NULL,
    sistema     boolean NOT NULL DEFAULT FALSE,
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW()
);

-- Roles de sistema: codigo unico globalmente (tenant_id IS NULL).
CREATE UNIQUE INDEX IF NOT EXISTS roles_sistema_codigo
    ON roles(codigo) WHERE tenant_id IS NULL;

-- Roles custom: codigo unico dentro do tenant.
CREATE UNIQUE INDEX IF NOT EXISTS roles_tenant_codigo
    ON roles(tenant_id, codigo) WHERE tenant_id IS NOT NULL;

DROP TRIGGER IF EXISTS roles_set_updated_at ON roles;
CREATE TRIGGER roles_set_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS permissions (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    codigo      varchar(80) NOT NULL UNIQUE,
    recurso     varchar(40) NOT NULL,
    acao        varchar(20) NOT NULL,
    descricao   varchar(200) NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id        uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id  uuid NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- =============================================================================
-- Cadastros base
-- =============================================================================
CREATE TABLE IF NOT EXISTS clientes (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id           uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    tipo                cliente_tipo NOT NULL,
    nome_razao_social   varchar(200) NOT NULL,
    nome_fantasia       varchar(200),
    documento           varchar(20) NOT NULL,
    email               citext,
    telefone_fixo       varchar(30),
    telefone_celular    varchar(30),
    contato_cliente     varchar(200),
    logradouro          varchar(200),
    numero              varchar(20),
    complemento         varchar(120),
    bairro              varchar(120),
    cidade              varchar(120),
    uf                  char(2),
    cep                 varchar(10),
    bloqueado           boolean NOT NULL DEFAULT FALSE,
    motivo_bloqueio     varchar(240),
    observacoes         text,
    created_at          timestamptz NOT NULL DEFAULT NOW(),
    updated_at          timestamptz NOT NULL DEFAULT NOW(),
    deleted_at          timestamptz,
    UNIQUE (tenant_id, documento)
);

CREATE INDEX IF NOT EXISTS idx_clientes_nome   ON clientes (tenant_id, lower(nome_razao_social)) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_clientes_doc    ON clientes (tenant_id, documento) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_clientes_cidade ON clientes (tenant_id, uf, cidade);

DROP TRIGGER IF EXISTS clientes_set_updated_at ON clientes;
CREATE TRIGGER clientes_set_updated_at BEFORE UPDATE ON clientes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS funcionarios (
    id                    uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id             uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    nome                  varchar(200) NOT NULL,
    documento             varchar(20),
    data_nascimento       date,
    data_admissao         date,
    data_desligamento     date,
    motivo_desligamento   varchar(240),
    cargo                 varchar(80),
    centro_custo          varchar(80),
    telefone              varchar(30),
    email                 citext,
    logradouro            varchar(200),
    numero                varchar(20),
    bairro                varchar(120),
    cidade                varchar(120),
    uf                    char(2),
    cep                   varchar(10),
    status                funcionario_status NOT NULL DEFAULT 'ativo',
    created_at            timestamptz NOT NULL DEFAULT NOW(),
    updated_at            timestamptz NOT NULL DEFAULT NOW(),
    deleted_at            timestamptz,
    UNIQUE (tenant_id, documento)
);

CREATE INDEX IF NOT EXISTS idx_func_tenant ON funcionarios(tenant_id) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS funcionarios_set_updated_at ON funcionarios;
CREATE TRIGGER funcionarios_set_updated_at BEFORE UPDATE ON funcionarios
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS usuarios (
    id              uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id       uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    funcionario_id  uuid REFERENCES funcionarios(id) ON DELETE SET NULL,
    -- login: parte apos "/" no formato "slug/login"
    login           varchar(60) NOT NULL,
    email           citext NOT NULL,
    senha_hash      varchar(120) NOT NULL,
    nome            varchar(200) NOT NULL,
    papel           usuario_papel NOT NULL DEFAULT 'comercial',
    ativo           boolean NOT NULL DEFAULT TRUE,
    ultimo_acesso   timestamptz,
    created_at      timestamptz NOT NULL DEFAULT NOW(),
    updated_at      timestamptz NOT NULL DEFAULT NOW(),
    deleted_at      timestamptz,
    UNIQUE (tenant_id, login),
    UNIQUE (tenant_id, email)
);

CREATE INDEX IF NOT EXISTS idx_usuarios_tenant ON usuarios(tenant_id) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS usuarios_set_updated_at ON usuarios;
CREATE TRIGGER usuarios_set_updated_at BEFORE UPDATE ON usuarios
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS user_roles (
    user_id     uuid NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    role_id     uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at  timestamptz NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Sessoes para BFF web (cookie -> session_id) e refresh do mobile.
CREATE TABLE IF NOT EXISTS user_sessions (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id      uuid NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    token_hash   varchar(64) NOT NULL UNIQUE,
    user_agent   varchar(240),
    ip           inet,
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    expires_at   timestamptz NOT NULL,
    revoked_at   timestamptz
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user    ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at);

-- =============================================================================
-- Frota
-- =============================================================================
CREATE TABLE IF NOT EXISTS veiculos (
    id                       uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id                uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    placa                    varchar(10) NOT NULL,
    descricao                varchar(120),
    marca                    varchar(60),
    modelo                   varchar(60),
    ano_fabricacao           smallint,
    ano_modelo               smallint,
    chassi                   varchar(30),
    renavam                  varchar(20),
    combustivel              varchar(20),
    cnpj_proprietario        varchar(20),
    numero_apolice_vigente   varchar(40),
    data_aquisicao           date,
    km_atual                 numeric(10,1) NOT NULL DEFAULT 0,
    status                   veiculo_status NOT NULL DEFAULT 'disponivel',
    created_at               timestamptz NOT NULL DEFAULT NOW(),
    updated_at               timestamptz NOT NULL DEFAULT NOW(),
    deleted_at               timestamptz,
    UNIQUE (tenant_id, placa)
);

DROP TRIGGER IF EXISTS veiculos_set_updated_at ON veiculos;
CREATE TRIGGER veiculos_set_updated_at BEFORE UPDATE ON veiculos
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- =============================================================================
-- Catalogo de servicos e equipamentos
-- =============================================================================
CREATE TABLE IF NOT EXISTS classificacoes_servico (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    codigo      varchar(40) NOT NULL,
    descricao   varchar(120) NOT NULL,
    ordem       int NOT NULL DEFAULT 0,
    ativo       boolean NOT NULL DEFAULT TRUE,
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, codigo)
);

DROP TRIGGER IF EXISTS class_serv_set_updated_at ON classificacoes_servico;
CREATE TRIGGER class_serv_set_updated_at BEFORE UPDATE ON classificacoes_servico
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS servicos_locacao (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id           uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    classificacao_id    uuid REFERENCES classificacoes_servico(id) ON DELETE SET NULL,
    codigo              varchar(40) NOT NULL,
    descricao           varchar(200) NOT NULL,
    unidade_padrao      varchar(20) NOT NULL DEFAULT 'DIARIA',
    valor_referencia    numeric(12,2) NOT NULL DEFAULT 0,
    ativo               boolean NOT NULL DEFAULT TRUE,
    created_at          timestamptz NOT NULL DEFAULT NOW(),
    updated_at          timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, codigo)
);

DROP TRIGGER IF EXISTS servicos_set_updated_at ON servicos_locacao;
CREATE TRIGGER servicos_set_updated_at BEFORE UPDATE ON servicos_locacao
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS equipamentos (
    id                       uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id                uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    codigo_patrimonio        varchar(40) NOT NULL,
    descricao                varchar(200) NOT NULL,
    marca                    varchar(80),
    modelo                   varchar(80),
    numero_serie             varchar(80),
    valor_aquisicao          numeric(12,2) NOT NULL DEFAULT 0,
    data_aquisicao           date,
    quantidade_total         int NOT NULL DEFAULT 1 CHECK (quantidade_total >= 0),
    quantidade_disponivel    int NOT NULL DEFAULT 1 CHECK (quantidade_disponivel >= 0),
    status                   equipamento_status NOT NULL DEFAULT 'disponivel',
    observacoes              text,
    created_at               timestamptz NOT NULL DEFAULT NOW(),
    updated_at               timestamptz NOT NULL DEFAULT NOW(),
    deleted_at               timestamptz,
    UNIQUE (tenant_id, codigo_patrimonio)
);

DROP TRIGGER IF EXISTS equip_set_updated_at ON equipamentos;
CREATE TRIGGER equip_set_updated_at BEFORE UPDATE ON equipamentos
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS kit_composicao (
    id                    uuid PRIMARY KEY DEFAULT uuidv7(),
    servico_locacao_id    uuid NOT NULL REFERENCES servicos_locacao(id) ON DELETE CASCADE,
    equipamento_id        uuid NOT NULL REFERENCES equipamentos(id) ON DELETE RESTRICT,
    usuario_cadastro_id   uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    quantidade            int NOT NULL DEFAULT 1 CHECK (quantidade > 0),
    created_at            timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (servico_locacao_id, equipamento_id)
);

-- =============================================================================
-- Motivos de cancelamento
-- =============================================================================
CREATE TABLE IF NOT EXISTS motivos_cancelamento (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id   uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    codigo      varchar(40) NOT NULL,
    descricao   varchar(200) NOT NULL,
    ativo       boolean NOT NULL DEFAULT TRUE,
    UNIQUE (tenant_id, codigo)
);

-- =============================================================================
-- Agenda (nucleo operacional)
-- =============================================================================
CREATE TABLE IF NOT EXISTS agenda (
    id                          uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id                   uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    -- numero sequencial por tenant, gerado por trigger
    numero                      bigint NOT NULL DEFAULT 0,
    cliente_id                  uuid NOT NULL REFERENCES clientes(id) ON DELETE RESTRICT,
    usuario_registro_id         uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    usuario_aprovador_id        uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    quem_contratou_id           uuid REFERENCES funcionarios(id) ON DELETE SET NULL,
    motivo_cancelamento_id      uuid REFERENCES motivos_cancelamento(id) ON DELETE SET NULL,
    status                      agenda_status NOT NULL DEFAULT 'orcamento',
    tipo_evento                 agenda_tipo_evento NOT NULL DEFAULT 'particular',
    tipo_retorno                agenda_tipo_retorno NOT NULL DEFAULT 'mesma_equipe',
    descricao_evento            varchar(240),
    data_evento                 date,
    hora_evento                 time,
    data_instalacao             date,
    hora_instalacao             time,
    data_retorno_prevista       date,
    data_retorno_real           date,
    forma_pagamento             forma_pagamento,
    valor_total                 numeric(12,2) NOT NULL DEFAULT 0,
    valor_desconto              numeric(12,2) NOT NULL DEFAULT 0,
    valor_liquido               numeric(12,2) NOT NULL DEFAULT 0,
    numero_aprovacao            varchar(40),
    data_aprovacao              timestamptz,
    data_cancelamento           timestamptz,
    finalizado                  boolean NOT NULL DEFAULT FALSE,
    observacoes                 text,
    created_at                  timestamptz NOT NULL DEFAULT NOW(),
    updated_at                  timestamptz NOT NULL DEFAULT NOW(),
    deleted_at                  timestamptz,
    UNIQUE (tenant_id, numero)
);

DROP TRIGGER IF EXISTS agenda_numero_trigger ON agenda;
CREATE TRIGGER agenda_numero_trigger BEFORE INSERT ON agenda
    FOR EACH ROW EXECUTE FUNCTION agenda_before_insert();

CREATE INDEX IF NOT EXISTS idx_agenda_tenant    ON agenda(tenant_id);
CREATE INDEX IF NOT EXISTS idx_agenda_cliente   ON agenda(cliente_id);
CREATE INDEX IF NOT EXISTS idx_agenda_status    ON agenda(tenant_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_agenda_evento_dt ON agenda(tenant_id, data_evento) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_agenda_numero    ON agenda(tenant_id, numero DESC) WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS agenda_set_updated_at ON agenda;
CREATE TRIGGER agenda_set_updated_at BEFORE UPDATE ON agenda
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS agenda_locais (
    id                       uuid PRIMARY KEY DEFAULT uuidv7(),
    agenda_id                uuid NOT NULL REFERENCES agenda(id) ON DELETE CASCADE,
    tipo                     agenda_local_tipo NOT NULL DEFAULT 'principal',
    apelido                  varchar(120),
    logradouro               varchar(200),
    numero                   varchar(20),
    complemento              varchar(120),
    bairro                   varchar(120),
    cidade                   varchar(120),
    uf                       char(2),
    cep                      varchar(10),
    ponto_referencia         varchar(200),
    localizacao_latitude     numeric(9,6),
    localizacao_longitude    numeric(9,6),
    distancia_km             numeric(8,2),
    principal                boolean NOT NULL DEFAULT FALSE,
    ordem                    int NOT NULL DEFAULT 0,
    observacoes              text,
    created_at               timestamptz NOT NULL DEFAULT NOW(),
    updated_at               timestamptz NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_agenda_locais_principal
    ON agenda_locais(agenda_id) WHERE principal;
CREATE INDEX IF NOT EXISTS idx_agenda_locais_agenda ON agenda_locais(agenda_id);

DROP TRIGGER IF EXISTS agenda_locais_set_updated_at ON agenda_locais;
CREATE TRIGGER agenda_locais_set_updated_at BEFORE UPDATE ON agenda_locais
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS agenda_local_contatos (
    id                       uuid PRIMARY KEY DEFAULT uuidv7(),
    agenda_local_id          uuid NOT NULL REFERENCES agenda_locais(id) ON DELETE CASCADE,
    nome                     varchar(200) NOT NULL,
    cargo                    varchar(80),
    telefone_principal       varchar(30),
    telefone_secundario      varchar(30),
    email                    citext,
    principal                boolean NOT NULL DEFAULT FALSE,
    observacoes              text,
    created_at               timestamptz NOT NULL DEFAULT NOW(),
    updated_at               timestamptz NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_agenda_local_contatos_principal
    ON agenda_local_contatos(agenda_local_id) WHERE principal;

DROP TRIGGER IF EXISTS agenda_loc_contatos_set_updated_at ON agenda_local_contatos;
CREATE TRIGGER agenda_loc_contatos_set_updated_at BEFORE UPDATE ON agenda_local_contatos
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS agenda_itens (
    id                       uuid PRIMARY KEY DEFAULT uuidv7(),
    agenda_id                uuid NOT NULL REFERENCES agenda(id) ON DELETE CASCADE,
    servico_locacao_id       uuid REFERENCES servicos_locacao(id) ON DELETE RESTRICT,
    numero_sequencial        int NOT NULL DEFAULT 0,
    descricao_complemento    text,
    quantidade               numeric(12,2) NOT NULL DEFAULT 1 CHECK (quantidade > 0),
    unidade                  varchar(20) NOT NULL DEFAULT 'DIARIA',
    valor_unitario           numeric(12,2) NOT NULL DEFAULT 0,
    valor_desconto           numeric(12,2) NOT NULL DEFAULT 0,
    valor_total              numeric(12,2) NOT NULL DEFAULT 0,
    observacoes              text
);

CREATE INDEX IF NOT EXISTS idx_agenda_itens_agenda ON agenda_itens(agenda_id);

CREATE TABLE IF NOT EXISTS agenda_equipe (
    id              uuid PRIMARY KEY DEFAULT uuidv7(),
    agenda_id       uuid NOT NULL REFERENCES agenda(id) ON DELETE CASCADE,
    funcionario_id  uuid NOT NULL REFERENCES funcionarios(id) ON DELETE RESTRICT,
    papel           agenda_equipe_papel NOT NULL DEFAULT 'instalacao',
    data_inicio     timestamptz,
    data_fim        timestamptz,
    observacoes     text,
    UNIQUE (agenda_id, funcionario_id, papel)
);

CREATE INDEX IF NOT EXISTS idx_agenda_equipe_func ON agenda_equipe(funcionario_id);

CREATE TABLE IF NOT EXISTS agenda_veiculos (
    id              uuid PRIMARY KEY DEFAULT uuidv7(),
    agenda_id       uuid NOT NULL REFERENCES agenda(id) ON DELETE CASCADE,
    veiculo_id      uuid NOT NULL REFERENCES veiculos(id) ON DELETE RESTRICT,
    motorista_id    uuid REFERENCES funcionarios(id) ON DELETE SET NULL,
    km_saida        numeric(10,1),
    km_retorno      numeric(10,1),
    data_saida      timestamptz,
    data_retorno    timestamptz,
    observacoes     text
);

CREATE INDEX IF NOT EXISTS idx_agenda_veiculos_veic ON agenda_veiculos(veiculo_id);

CREATE TABLE IF NOT EXISTS agenda_status_historico (
    id                uuid PRIMARY KEY DEFAULT uuidv7(),
    agenda_id         uuid NOT NULL REFERENCES agenda(id) ON DELETE CASCADE,
    usuario_id        uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    status_anterior   agenda_status,
    status_novo       agenda_status NOT NULL,
    motivo            text,
    ocorreu_em        timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agenda_status_hist ON agenda_status_historico(agenda_id, ocorreu_em DESC);

-- =============================================================================
-- Estoque (livro-razao)
-- =============================================================================
CREATE TABLE IF NOT EXISTS movimentacoes_estoque (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    agenda_id           uuid REFERENCES agenda(id) ON DELETE SET NULL,
    equipamento_id      uuid NOT NULL REFERENCES equipamentos(id) ON DELETE RESTRICT,
    responsavel_id      uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    tipo                movimentacao_tipo NOT NULL,
    quantidade          int NOT NULL CHECK (quantidade <> 0),
    data_movimentacao   timestamptz NOT NULL DEFAULT NOW(),
    observacoes         text
);

CREATE INDEX IF NOT EXISTS idx_mov_estoque_equip   ON movimentacoes_estoque(equipamento_id);
CREATE INDEX IF NOT EXISTS idx_mov_estoque_agenda  ON movimentacoes_estoque(agenda_id);
CREATE INDEX IF NOT EXISTS idx_mov_estoque_data    ON movimentacoes_estoque(data_movimentacao DESC);

-- =============================================================================
-- Financeiro
-- =============================================================================
CREATE TABLE IF NOT EXISTS contas_receber (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id           uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    agenda_id           uuid REFERENCES agenda(id) ON DELETE SET NULL,
    cliente_id          uuid NOT NULL REFERENCES clientes(id) ON DELETE RESTRICT,
    -- numero_titulo gerado por trigger: 'CR-' || seq por tenant
    numero_titulo       varchar(40) NOT NULL DEFAULT '',
    competencia         char(7) NOT NULL,
    data_emissao        date NOT NULL DEFAULT CURRENT_DATE,
    data_vencimento     date NOT NULL,
    valor_original      numeric(12,2) NOT NULL DEFAULT 0,
    valor_baixado       numeric(12,2) NOT NULL DEFAULT 0,
    saldo               numeric(12,2) NOT NULL DEFAULT 0,
    status              conta_status NOT NULL DEFAULT 'em_aberto',
    observacoes         text,
    created_at          timestamptz NOT NULL DEFAULT NOW(),
    updated_at          timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, numero_titulo)
);

DROP TRIGGER IF EXISTS contas_receber_numero_trigger ON contas_receber;
CREATE TRIGGER contas_receber_numero_trigger BEFORE INSERT ON contas_receber
    FOR EACH ROW EXECUTE FUNCTION contas_receber_before_insert();

CREATE INDEX IF NOT EXISTS idx_cr_tenant    ON contas_receber(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cr_cliente   ON contas_receber(cliente_id);
CREATE INDEX IF NOT EXISTS idx_cr_status    ON contas_receber(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_cr_vencto    ON contas_receber(tenant_id, data_vencimento);

DROP TRIGGER IF EXISTS contas_receber_set_updated_at ON contas_receber;
CREATE TRIGGER contas_receber_set_updated_at BEFORE UPDATE ON contas_receber
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS recebimentos (
    id                      uuid PRIMARY KEY DEFAULT uuidv7(),
    conta_receber_id        uuid NOT NULL REFERENCES contas_receber(id) ON DELETE CASCADE,
    usuario_registro_id     uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    data_recebimento        date NOT NULL DEFAULT CURRENT_DATE,
    valor_recebido          numeric(12,2) NOT NULL CHECK (valor_recebido <> 0),
    forma_pagamento         forma_pagamento NOT NULL,
    tipo_documento          tipo_documento_fiscal,
    numero_documento        varchar(60),
    referencia              varchar(120),
    observacoes             text,
    created_at              timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recebimentos_conta ON recebimentos(conta_receber_id);

-- =============================================================================
-- Anexos (polimorfico), parametros, auditoria
-- =============================================================================
CREATE TABLE IF NOT EXISTS anexos (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id           uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    usuario_envio_id    uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    entidade            varchar(40) NOT NULL,
    registro_id         uuid NOT NULL,
    nome_arquivo        varchar(240) NOT NULL,
    url                 varchar(500) NOT NULL,
    tipo_mime           varchar(120),
    tamanho_bytes       bigint,
    descricao           text,
    created_at          timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_anexos_tenant ON anexos(tenant_id);
CREATE INDEX IF NOT EXISTS idx_anexos_target ON anexos(entidade, registro_id);

CREATE TABLE IF NOT EXISTS parametros_sistema (
    id                          uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id                   uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    usuario_atualizacao_id      uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    chave                       varchar(80) NOT NULL,
    valor                       text,
    descricao                   text,
    updated_at                  timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, chave)
);

DROP TRIGGER IF EXISTS parametros_set_updated_at ON parametros_sistema;
CREATE TRIGGER parametros_set_updated_at BEFORE UPDATE ON parametros_sistema
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS logs_auditoria (
    id              uuid PRIMARY KEY DEFAULT uuidv7(),
    tenant_id       uuid REFERENCES tenants(id) ON DELETE SET NULL,
    usuario_id      uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    entidade        varchar(40) NOT NULL,
    registro_id     uuid,
    acao            varchar(40) NOT NULL,
    valor_anterior  jsonb,
    valor_novo      jsonb,
    ip_origem       inet,
    user_agent      varchar(240),
    ocorreu_em      timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_tenant  ON logs_auditoria(tenant_id, ocorreu_em DESC);
CREATE INDEX IF NOT EXISTS idx_audit_target  ON logs_auditoria(entidade, registro_id);
CREATE INDEX IF NOT EXISTS idx_audit_usuario ON logs_auditoria(usuario_id, ocorreu_em DESC);

-- =============================================================================
-- Seed RBAC — roles e permissions de sistema (tenant_id IS NULL)
-- =============================================================================
INSERT INTO roles (codigo, descricao, sistema, tenant_id) VALUES
    ('admin',      'Administrador geral',                  TRUE, NULL),
    ('comercial',  'Gestao de clientes e orcamentos',      TRUE, NULL),
    ('operacao',   'Agenda, equipes e execucao',           TRUE, NULL),
    ('financeiro', 'Contas a receber e baixas',            TRUE, NULL),
    ('fiscal',     'Documentos fiscais e auditoria',       TRUE, NULL),
    ('campo',      'Operacao em campo (mobile)',           TRUE, NULL)
ON CONFLICT DO NOTHING;

INSERT INTO permissions (codigo, recurso, acao, descricao) VALUES
    ('clientes.read',          'clientes',          'read',   'Visualizar clientes'),
    ('clientes.write',         'clientes',          'write',  'Criar/editar clientes'),
    ('agenda.read',            'agenda',            'read',   'Visualizar agenda e OS'),
    ('agenda.write',           'agenda',            'write',  'Criar/editar agenda e OS'),
    ('agenda.approve',         'agenda',            'approve','Aprovar OS'),
    ('agenda.cancel',          'agenda',            'cancel', 'Cancelar OS'),
    ('estoque.read',           'estoque',           'read',   'Visualizar estoque'),
    ('estoque.write',          'estoque',           'write',  'Movimentar estoque'),
    ('frota.read',             'frota',             'read',   'Visualizar frota'),
    ('frota.write',            'frota',             'write',  'Editar frota'),
    ('financeiro.read',        'financeiro',        'read',   'Visualizar financeiro'),
    ('financeiro.write',       'financeiro',        'write',  'Lancar baixas e contas'),
    ('fiscal.read',            'fiscal',            'read',   'Visualizar documentos fiscais'),
    ('fiscal.write',           'fiscal',            'write',  'Emitir/cancelar documentos'),
    ('rh.read',                'rh',                'read',   'Visualizar funcionarios'),
    ('rh.write',               'rh',                'write',  'Editar funcionarios'),
    ('admin.usuarios',         'admin',             'manage', 'Gerenciar usuarios e perfis'),
    ('admin.parametros',       'admin',             'manage', 'Gerenciar parametros'),
    ('admin.auditoria',        'admin',             'read',   'Visualizar logs de auditoria')
ON CONFLICT (codigo) DO NOTHING;

-- admin recebe TUDO
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r CROSS JOIN permissions p
WHERE r.codigo = 'admin' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

-- comercial
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.codigo IN (
    'clientes.read','clientes.write',
    'agenda.read','agenda.write',
    'estoque.read','frota.read','financeiro.read'
)
WHERE r.codigo = 'comercial' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

-- operacao
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.codigo IN (
    'agenda.read','agenda.write','agenda.approve','agenda.cancel',
    'estoque.read','estoque.write',
    'frota.read','frota.write',
    'rh.read','clientes.read'
)
WHERE r.codigo = 'operacao' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

-- financeiro
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.codigo IN (
    'financeiro.read','financeiro.write',
    'clientes.read','agenda.read','fiscal.read'
)
WHERE r.codigo = 'financeiro' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

-- fiscal
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.codigo IN (
    'fiscal.read','fiscal.write',
    'financeiro.read','agenda.read','clientes.read','admin.auditoria'
)
WHERE r.codigo = 'fiscal' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

-- campo (mobile)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.codigo IN (
    'agenda.read','estoque.read','estoque.write','frota.read'
)
WHERE r.codigo = 'campo' AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;
