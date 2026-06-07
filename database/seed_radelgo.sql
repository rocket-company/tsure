-- =============================================================================
-- seed_radelgo.sql  — Importação do primeiro tenant: Radelgo Sonorizacao e Eventos
--
-- Execute a partir da RAIZ do projeto:
--   psql -d <banco> -f database/seed_radelgo.sql
--
-- Pré-requisitos: schema.sql já aplicado.
-- Os CSVs devem estar em database-access/exports/ (relativo ao diretório de
-- execução do psql, ou seja, a raiz do projeto).
-- =============================================================================

SET client_min_messages = NOTICE;
\set ON_ERROR_STOP on

-- ---------------------------------------------------------------------------
-- 0. Guarda: aborta se o tenant ja existe
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM tenants WHERE slug = 'radelgo') THEN
        RAISE EXCEPTION
            'Tenant "radelgo" ja existe. '
            'Para reimportar, execute: '
            'DELETE FROM tenants WHERE slug = ''radelgo'' CASCADE;';
    END IF;
END $$;

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Tabelas temporarias para dados brutos dos CSVs (tudo text)
-- ---------------------------------------------------------------------------

CREATE TEMP TABLE _csv_class (
    "Código"   text,
    "DescClass" text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_loc (
    "IdLoc"   text,
    "DescLoc" text,
    "Clas"    text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_func (
    "Idfunc"     text,
    "NomeFunc"   text,
    "DtNascFunc" text,
    "DtAdmFunc"  text,
    "DesCargFunc" text,
    "StatusFunc" text,
    "CCusto"     text,
    "TelFunc"    text,
    "EndFunc"    text,
    "NrEndFunc"  text,
    "BairroFunc" text,
    "CidFunc"    text,
    "DtDesliFunc" text,
    "MotDeslFunc" text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_usuario (
    "Idus"       text,
    "IdUsuario"  text,
    "Nome"       text,
    "Senha"      text,
    "Tipo"       text,
    "DataCadastro" text,
    "IdDeptoExt" text,
    "IdSetorExt" text,
    "Ramal"      text,
    "DataDesl"   text,
    "CliforSap"  text,
    "IdCargoExt" text,
    "EndOutlook" text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_motcanc (
    "IdMotCanc"  text,
    "DescMotivo" text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_cliente (
    "IdCliente"      text,
    "RazCliente"     text,
    "Tipo"           text,
    "CNPJ_CPF"       text,
    "End"            text,
    "nr"             text,
    "Compl"          text,
    "Bairro"         text,
    "Cidade"         text,
    "ContatoCliente" text,
    "TelFixo"        text,
    "TelCel"         text,
    "ObsCliente"     text,
    "email"          text,
    "Blq"            text,
    "Motivo"         text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_veiculo (
    "IdVeic"   text,
    "DescVeic" text,
    "Placa"    text,
    "MarcaVeic" text,
    "DtAquis"  text,
    "Renavan"  text,
    "Chassi"   text,
    "AnoFab"   text,
    "Comb"     text,
    "CNPJ"     text,
    "NrApoVig" text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_equip (
    "IdEquip"   text,
    "DescEquip" text,
    "Marca"     text,
    "Vlr"       text,
    "SemUso"    text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_kit (
    "IdKit"          text,
    "IdItemLocExt"   text,
    "IdItemEquipExt" text,
    "QdeItem"        text,
    "IdUsCad"        text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_agenda (
    "IdServ"        text,
    "IdClient"      text,
    "DtAg"          text,
    "HrAg"          text,
    "IdUs"          text,
    "QuemNeg"       text,
    "Status"        text,
    "TipRet"        text,
    "FormPag"       text,
    "DtAprov"       text,
    "QuemAprov"     text,
    "NrAprov"       text,
    "DescEvento"    text,
    "EndEvento"     text,
    "NrEnd"         text,
    "ComplEnd"      text,
    "Bairro"        text,
    "Cidade"        text,
    "LocGPS"        text,
    "ContEvent"     text,
    "TelContEvent"  text,
    "Cont2Event"    text,
    "Tel2ContEvent" text,
    "DistEvent"     text,
    "TipEvent"      text,
    "DtEvent"       text,
    "HrEvent"       text,
    "datInst"       text,
    "HrInst"        text,
    "Obsagend"      text,
    "StAberto"      text,
    "DtCanc"        text,
    "QuemCanc"      text,
    "MotCanc"       text,
    "CancEvent"     text,
    "IdFuncInst"    text,
    "StInst"        text,
    "IdFunciRet"    text,
    "DtRet"         text,
    "DtRealRet"     text,
    "StConc"        text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_agenda_itens (
    "IdItensAgend" text,
    "IdAgendaExt"  text,
    "NrServ"       text,
    "UnidItem"     text,
    "IdItemExt"    text,
    "ComplItem"    text,
    "Qdade"        text,
    "Unid"         text,
    "VlrUnit"      text,
    "ObsItem"      text,
    "Desc"         text,
    "VlrTotal"     text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_escala (
    "IdEscala"  text,
    "IdServExt" text,
    "IdFuncExt" text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_escala_ret (
    "IdEscala"  text,
    "IdServExt" text,
    "IdFuncExt" text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_saida_veic (
    "IdSaidVeic" text,
    "IdEvento"   text,
    "IdVeicExt"  text,
    "KmSaida"    text,
    "KmRet"      text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_saida_veic_ret (
    "IdSaidVeic" text,
    "IdEvento"   text,
    "IdVeicExt"  text,
    "KmSaida"    text,
    "KmRet"      text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_equip_saida (
    "IdItemEquip" text,
    "IdAgendExt"  text,
    "IdItemExt"   text,
    "QdaSaida"    text
) ON COMMIT DROP;

CREATE TEMP TABLE _csv_pgto (
    "IdRec"      text,
    "IdAgendExt" text,
    "DtReg"      text,
    "IdUs"       text,
    "VlrRec"     text,
    "DtRec"      text,
    "FormRec"    text,
    "DocRec"     text,
    "NrDocRec"   text
) ON COMMIT DROP;

-- ---------------------------------------------------------------------------
-- 2. Carrega CSVs via \copy (caminhos relativos à raiz do projeto)
-- ---------------------------------------------------------------------------

\copy _csv_class     FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__CadClass.csv'      CSV HEADER ENCODING 'UTF8'
\copy _csv_loc       FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__CadLocacao.csv'    CSV HEADER ENCODING 'UTF8'
\copy _csv_func      FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__CadFunc.csv'       CSV HEADER ENCODING 'UTF8'
\copy _csv_usuario   FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__CadUsuario.csv'    CSV HEADER ENCODING 'UTF8'
\copy _csv_motcanc   FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabMotCanc.csv'    CSV HEADER ENCODING 'UTF8'
\copy _csv_cliente   FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabClientes.csv'   CSV HEADER ENCODING 'UTF8'
\copy _csv_veiculo   FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabVeiculo.csv'    CSV HEADER ENCODING 'UTF8'
\copy _csv_equip     FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabEquip.csv'      CSV HEADER ENCODING 'UTF8'
\copy _csv_kit       FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabKit.csv'        CSV HEADER ENCODING 'UTF8'
\copy _csv_agenda    FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabAgenda.csv'     CSV HEADER ENCODING 'UTF8'
\copy _csv_agenda_itens FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabAgendaItens.csv' CSV HEADER ENCODING 'UTF8'
\copy _csv_escala    FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabEscala.csv'     CSV HEADER ENCODING 'UTF8'
\copy _csv_escala_ret FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabEscalaRet.csv' CSV HEADER ENCODING 'UTF8'
\copy _csv_saida_veic FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabSaidaVeic.csv' CSV HEADER ENCODING 'UTF8'
\copy _csv_saida_veic_ret FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabSaidaVeicRet.csv' CSV HEADER ENCODING 'UTF8'
\copy _csv_equip_saida FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabEquipSaida.csv' CSV HEADER ENCODING 'UTF8'
\copy _csv_pgto      FROM 'database-access/exports/SISTEMA_GESTAO_RADELGO_TABELAS__TabPgto.csv'       CSV HEADER ENCODING 'UTF8'

-- ---------------------------------------------------------------------------
-- 3. Tenant Radelgo
-- ---------------------------------------------------------------------------

INSERT INTO tenants (slug, nome, plano, ativo)
VALUES ('radelgo', 'Radelgo Sonorizacao e Eventos', 'standard', TRUE);

-- Referencia ao tenant_id usada em todos os INSERTs abaixo
CREATE TEMP TABLE _t AS
    SELECT id AS tid FROM tenants WHERE slug = 'radelgo' LIMIT 1;

-- ---------------------------------------------------------------------------
-- 4. Mapeamento de IDs legados → UUIDs novos
-- ---------------------------------------------------------------------------

-- classificacoes_servico
CREATE TEMP TABLE _class_map (
    old_id  int PRIMARY KEY,
    new_id  uuid NOT NULL DEFAULT uuidv7()
) ON COMMIT DROP;

INSERT INTO _class_map (old_id)
SELECT CAST(TRIM("Código") AS int)
FROM _csv_class
WHERE NULLIF(TRIM("Código"), '') IS NOT NULL;

-- servicos_locacao
CREATE TEMP TABLE _loc_map (
    old_id  int PRIMARY KEY,
    new_id  uuid NOT NULL DEFAULT uuidv7()
) ON COMMIT DROP;

INSERT INTO _loc_map (old_id)
SELECT CAST(TRIM("IdLoc") AS int)
FROM _csv_loc
WHERE NULLIF(TRIM("IdLoc"), '') IS NOT NULL;

-- funcionarios
CREATE TEMP TABLE _func_map (
    old_id  int PRIMARY KEY,
    new_id  uuid NOT NULL DEFAULT uuidv7()
) ON COMMIT DROP;

INSERT INTO _func_map (old_id)
SELECT CAST(TRIM(CAST("Idfunc" AS float)::int::text) AS int)
FROM _csv_func
WHERE NULLIF(TRIM("NomeFunc"), '') IS NOT NULL;

-- usuarios
CREATE TEMP TABLE _us_map (
    old_id  int PRIMARY KEY,
    new_id  uuid NOT NULL DEFAULT uuidv7()
) ON COMMIT DROP;

INSERT INTO _us_map (old_id)
SELECT CAST(TRIM("Idus") AS int)
FROM _csv_usuario
WHERE NULLIF(TRIM("Idus"), '') IS NOT NULL;

-- motivos_cancelamento
CREATE TEMP TABLE _motcanc_map (
    old_id  int PRIMARY KEY,
    new_id  uuid NOT NULL DEFAULT uuidv7()
) ON COMMIT DROP;

INSERT INTO _motcanc_map (old_id)
SELECT CAST(TRIM("IdMotCanc") AS int)
FROM _csv_motcanc
WHERE NULLIF(TRIM("IdMotCanc"), '') IS NOT NULL;

-- clientes
CREATE TEMP TABLE _cliente_map (
    old_id  int PRIMARY KEY,
    new_id  uuid NOT NULL DEFAULT uuidv7()
) ON COMMIT DROP;

INSERT INTO _cliente_map (old_id)
SELECT CAST(CAST(TRIM("IdCliente") AS float)::int AS int)
FROM _csv_cliente
WHERE NULLIF(TRIM("IdCliente"), '') IS NOT NULL;

-- veiculos
CREATE TEMP TABLE _veic_map (
    old_id  int PRIMARY KEY,
    new_id  uuid NOT NULL DEFAULT uuidv7()
) ON COMMIT DROP;

INSERT INTO _veic_map (old_id)
SELECT CAST(TRIM("IdVeic") AS int)
FROM _csv_veiculo
WHERE NULLIF(TRIM("IdVeic"), '') IS NOT NULL;

-- equipamentos
CREATE TEMP TABLE _equip_map (
    old_id  int PRIMARY KEY,
    new_id  uuid NOT NULL DEFAULT uuidv7()
) ON COMMIT DROP;

INSERT INTO _equip_map (old_id)
SELECT CAST(CAST(TRIM("IdEquip") AS float)::int AS int)
FROM _csv_equip
WHERE NULLIF(TRIM("IdEquip"), '') IS NOT NULL;

-- agenda
CREATE TEMP TABLE _agenda_map (
    old_id  int PRIMARY KEY,
    new_id  uuid NOT NULL DEFAULT uuidv7()
) ON COMMIT DROP;

INSERT INTO _agenda_map (old_id)
SELECT CAST(TRIM("IdServ") AS int)
FROM _csv_agenda
WHERE NULLIF(TRIM("IdServ"), '') IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 5. classificacoes_servico
-- ---------------------------------------------------------------------------

INSERT INTO classificacoes_servico (id, tenant_id, codigo, descricao, ordem)
SELECT
    m.new_id,
    (SELECT tid FROM _t),
    TRIM(c."Código"),
    TRIM(c."DescClass"),
    CAST(TRIM(c."Código") AS int)
FROM _csv_class c
JOIN _class_map m ON CAST(TRIM(c."Código") AS int) = m.old_id;

-- ---------------------------------------------------------------------------
-- 6. servicos_locacao
-- ---------------------------------------------------------------------------

INSERT INTO servicos_locacao (id, tenant_id, classificacao_id, codigo, descricao)
SELECT
    lm.new_id,
    (SELECT tid FROM _t),
    cm.new_id,
    LPAD(TRIM(l."IdLoc"), 4, '0'),
    TRIM(l."DescLoc")
FROM _csv_loc l
JOIN _loc_map lm ON CAST(TRIM(l."IdLoc") AS int) = lm.old_id
-- Tenta mapear a classificação pelo nome (Clas = DescClass legado)
LEFT JOIN classificacoes_servico cs
    ON cs.tenant_id = (SELECT tid FROM _t)
    AND lower(cs.descricao) = lower(TRIM(l."Clas"))
LEFT JOIN _class_map cm ON cm.new_id = cs.id;

-- ---------------------------------------------------------------------------
-- 7. funcionarios
-- ---------------------------------------------------------------------------

INSERT INTO funcionarios (
    id, tenant_id, nome, data_nascimento, data_admissao, data_desligamento,
    motivo_desligamento, cargo, centro_custo, telefone,
    logradouro, numero, bairro, cidade,
    status, documento
)
SELECT
    m.new_id,
    (SELECT tid FROM _t),
    TRIM(f."NomeFunc"),
    NULLIF(TRIM(f."DtNascFunc"), '')::date,
    NULLIF(TRIM(f."DtAdmFunc"), '')::date,
    NULLIF(TRIM(f."DtDesliFunc"), '')::date,
    NULLIF(TRIM(f."MotDeslFunc"), ''),
    NULLIF(TRIM(f."DesCargFunc"), ''),
    NULLIF(TRIM(f."CCusto"), ''),
    NULLIF(TRIM(f."TelFunc"), ''),
    NULLIF(TRIM(f."EndFunc"), ''),
    NULLIF(TRIM(f."NrEndFunc"), ''),
    NULLIF(TRIM(f."BairroFunc"), ''),
    NULLIF(TRIM(f."CidFunc"), ''),
    CASE
        WHEN lower(TRIM(f."StatusFunc")) = 'ativo'      THEN 'ativo'::funcionario_status
        WHEN lower(TRIM(f."StatusFunc")) = 'desligado'  THEN 'desligado'::funcionario_status
        ELSE 'afastado'::funcionario_status
    END,
    -- CPF/doc nao existe no legado; placeholder para preservar unicidade
    'LEGADO-' || LPAD(CAST(CAST(TRIM(f."Idfunc") AS float)::int AS text), 6, '0')
FROM _csv_func f
JOIN _func_map m ON CAST(CAST(TRIM(f."Idfunc") AS float)::int AS int) = m.old_id
WHERE NULLIF(TRIM(f."NomeFunc"), '') IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 8. usuarios
-- ---------------------------------------------------------------------------

INSERT INTO usuarios (
    id, tenant_id, login, email, senha_hash, nome, papel, ativo
)
SELECT
    um.new_id,
    (SELECT tid FROM _t),
    lower(TRIM(u."IdUsuario")),
    -- Email nao existe no CSV; placeholder derivado do login
    lower(TRIM(u."IdUsuario")) || '@radelgo.local',
    crypt(
        COALESCE(NULLIF(TRIM(u."Senha"), ''), 'radelgo@' || TRIM(u."Idus")),
        gen_salt('bf')
    ),
    initcap(lower(TRIM(u."Nome"))),
    CASE upper(TRIM(u."Tipo"))
        WHEN 'ADM' THEN 'admin'::usuario_papel
        ELSE            'comercial'::usuario_papel
    END,
    (NULLIF(TRIM(u."DataDesl"), '') IS NULL)
FROM _csv_usuario u
JOIN _us_map um ON CAST(TRIM(u."Idus") AS int) = um.old_id
WHERE NULLIF(TRIM(u."Idus"), '') IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 9. motivos_cancelamento
-- ---------------------------------------------------------------------------

INSERT INTO motivos_cancelamento (id, tenant_id, codigo, descricao)
SELECT
    m.new_id,
    (SELECT tid FROM _t),
    LPAD(TRIM(c."IdMotCanc"), 4, '0'),
    TRIM(c."DescMotivo")
FROM _csv_motcanc c
JOIN _motcanc_map m ON CAST(TRIM(c."IdMotCanc") AS int) = m.old_id
WHERE NULLIF(TRIM(c."IdMotCanc"), '') IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 10. clientes  (documento pode ser duplicado: ON CONFLICT DO NOTHING)
-- ---------------------------------------------------------------------------

-- Remove caracteres nao numericos do documento para deduplica
INSERT INTO clientes (
    id, tenant_id, tipo, nome_razao_social, documento,
    logradouro, numero, complemento, bairro, cidade,
    telefone_fixo, telefone_celular, contato_cliente,
    email, bloqueado, motivo_bloqueio, observacoes
)
SELECT DISTINCT ON (
    REGEXP_REPLACE(COALESCE(NULLIF(TRIM(c."CNPJ_CPF"), ''), 'SEM-DOC'), '[^0-9]', '', 'g')
)
    cm.new_id,
    (SELECT tid FROM _t),
    CASE
        WHEN TRIM(c."Tipo") ILIKE '%sic%' OR TRIM(c."Tipo") ILIKE '%fica%'
            THEN 'pessoa_fisica'::cliente_tipo
        ELSE    'pessoa_juridica'::cliente_tipo
    END,
    COALESCE(NULLIF(TRIM(c."RazCliente"), ''), 'Cliente ' || LPAD(CAST(CAST(TRIM(c."IdCliente") AS float)::int AS text), 6, '0')),
    COALESCE(
        NULLIF(REGEXP_REPLACE(TRIM(c."CNPJ_CPF"), '[^0-9]', '', 'g'), ''),
        'SEM-DOC-' || LPAD(CAST(CAST(TRIM(c."IdCliente") AS float)::int AS text), 6, '0')
    ),
    NULLIF(TRIM(c."End"), ''),
    NULLIF(TRIM(c."nr"), ''),
    NULLIF(TRIM(c."Compl"), ''),
    NULLIF(TRIM(c."Bairro"), ''),
    NULLIF(TRIM(c."Cidade"), ''),
    NULLIF(TRIM(c."TelFixo"), ''),
    NULLIF(TRIM(c."TelCel"), ''),
    NULLIF(TRIM(c."ContatoCliente"), ''),
    NULLIF(TRIM(c."email"), ''),
    (TRIM(c."Blq") = 'True'),
    NULLIF(TRIM(c."Motivo"), ''),
    NULLIF(TRIM(c."ObsCliente"), '')
FROM _csv_cliente c
JOIN _cliente_map cm ON CAST(CAST(TRIM(c."IdCliente") AS float)::int AS int) = cm.old_id
WHERE NULLIF(TRIM(c."IdCliente"), '') IS NOT NULL
ORDER BY
    REGEXP_REPLACE(COALESCE(NULLIF(TRIM(c."CNPJ_CPF"), ''), 'SEM-DOC'), '[^0-9]', '', 'g'),
    cm.new_id;

-- Atualiza o mapeamento para clientes que foram ignorados por duplicata de documento
-- (mantém o primeiro uuid inserido)
CREATE TEMP TABLE _cliente_doc_map AS
SELECT
    cm.old_id,
    COALESCE(
        (SELECT cl.id
         FROM clientes cl
         WHERE cl.tenant_id = (SELECT tid FROM _t)
           AND cl.documento = COALESCE(
               NULLIF(REGEXP_REPLACE(TRIM(c."CNPJ_CPF"), '[^0-9]', '', 'g'), ''),
               'SEM-DOC-' || LPAD(CAST(CAST(TRIM(c."IdCliente") AS float)::int AS text), 6, '0')
           )
         LIMIT 1),
        cm.new_id
    ) AS resolved_id
FROM _csv_cliente c
JOIN _cliente_map cm ON CAST(CAST(TRIM(c."IdCliente") AS float)::int AS int) = cm.old_id
WHERE NULLIF(TRIM(c."IdCliente"), '') IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 11. veiculos
-- ---------------------------------------------------------------------------

INSERT INTO veiculos (
    id, tenant_id, placa, descricao, marca,
    ano_fabricacao, chassi, renavam, combustivel,
    cnpj_proprietario, numero_apolice_vigente, data_aquisicao
)
SELECT
    vm.new_id,
    (SELECT tid FROM _t),
    -- Placa: usa placeholder se vazio (ex: veículo terceirizado)
    COALESCE(
        NULLIF(UPPER(TRIM(v."Placa")), ''),
        'TERC-' || LPAD(TRIM(v."IdVeic"), 4, '0')
    ),
    NULLIF(TRIM(v."DescVeic"), ''),
    NULLIF(TRIM(v."MarcaVeic"), ''),
    -- AnoFab pode ser "1983/1983" — extrai os 4 primeiros chars
    NULLIF(SUBSTRING(TRIM(v."AnoFab") FROM 1 FOR 4), '')::smallint,
    NULLIF(TRIM(v."Chassi"), ''),
    NULLIF(TRIM(v."Renavan"), ''),
    NULLIF(TRIM(TRIM(TRAILING ',' FROM v."Comb")), ''),
    NULLIF(TRIM(v."CNPJ"), ''),
    NULLIF(TRIM(v."NrApoVig"), ''),
    NULLIF(TRIM(v."DtAquis"), '')::date
FROM _csv_veiculo v
JOIN _veic_map vm ON CAST(TRIM(v."IdVeic") AS int) = vm.old_id
WHERE NULLIF(TRIM(v."IdVeic"), '') IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 12. equipamentos
-- ---------------------------------------------------------------------------

INSERT INTO equipamentos (
    id, tenant_id, codigo_patrimonio, descricao, marca,
    valor_aquisicao, status
)
SELECT
    em.new_id,
    (SELECT tid FROM _t),
    -- Usa o Id legado como codigo de patrimônio
    'PAT-' || LPAD(CAST(CAST(TRIM(e."IdEquip") AS float)::int AS text), 6, '0'),
    COALESCE(NULLIF(TRIM(e."DescEquip"), ''), 'Equipamento ' || TRIM(e."IdEquip")),
    NULLIF(TRIM(e."Marca"), ''),
    COALESCE(NULLIF(TRIM(e."Vlr"), '')::numeric, 0),
    CASE
        WHEN TRIM(e."SemUso") = 'True' THEN 'em_manutencao'::equipamento_status
        ELSE 'disponivel'::equipamento_status
    END
FROM _csv_equip e
JOIN _equip_map em ON CAST(CAST(TRIM(e."IdEquip") AS float)::int AS int) = em.old_id
WHERE NULLIF(TRIM(e."IdEquip"), '') IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 13. kit_composicao
-- ---------------------------------------------------------------------------

INSERT INTO kit_composicao (servico_locacao_id, equipamento_id, usuario_cadastro_id, quantidade)
SELECT
    lm.new_id,
    eqm.new_id,
    usm.new_id,
    COALESCE(NULLIF(TRIM(k."QdeItem"), '')::int, 1)
FROM _csv_kit k
JOIN _loc_map  lm  ON CAST(TRIM(k."IdItemLocExt") AS int)           = lm.old_id
JOIN _equip_map eqm ON CAST(CAST(TRIM(k."IdItemEquipExt") AS float)::int AS int) = eqm.old_id
LEFT JOIN _us_map usm ON CAST(TRIM(k."IdUsCad") AS int)             = usm.old_id
WHERE NULLIF(TRIM(k."IdItemLocExt"), '') IS NOT NULL
  AND NULLIF(TRIM(k."IdItemEquipExt"), '') IS NOT NULL
ON CONFLICT (servico_locacao_id, equipamento_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 14. agenda
--
-- Mapeamento de status legado:
--   DtCanc preenchida (nao ficticia)         → cancelado
--   "Ordem Serviço" + StConc=True            → finalizado
--   "Ordem Serviço" + StInst=True            → aguardando_retorno
--   "Ordem Serviço"                          → agendado
--   "Orçamento" / qualquer outro             → orcamento
--
-- Datas fictícias (1899-12-30, 2025-01-09):
--   HrAg / HrEvent / HrInst: extrai apenas parte time do valor bogus
--   DtCanc = 2025-01-09 é artefato do sistema → NULL
-- ---------------------------------------------------------------------------

INSERT INTO agenda (
    id, tenant_id, numero, cliente_id, usuario_registro_id,
    motivo_cancelamento_id,
    status, tipo_evento, tipo_retorno, forma_pagamento,
    descricao_evento, data_evento, hora_evento,
    data_instalacao, hora_instalacao,
    data_retorno_prevista, data_retorno_real,
    numero_aprovacao, data_aprovacao, data_cancelamento,
    finalizado, observacoes, created_at
)
SELECT
    am.new_id,
    (SELECT tid FROM _t),
    -- numero: será sobrescrito pelo trigger, mas passamos 0 explicitamente
    0,
    -- cliente_id: resolve via mapa de deduplicacao por documento
    cdm.resolved_id,
    -- usuario_registro_id
    usm.new_id,
    -- motivo_cancelamento_id
    mcm.new_id,
    -- status derivado
    CASE
        WHEN NULLIF(TRIM(a."DtCanc"), '') IS NOT NULL
          AND TRIM(a."DtCanc") NOT LIKE '2025%'
          AND TRIM(a."DtCanc") NOT LIKE '%0001%'
        THEN 'cancelado'::agenda_status
        WHEN TRIM(a."Status") = 'Ordem Serviço' AND a."StConc" = 'True'
        THEN 'finalizado'::agenda_status
        WHEN TRIM(a."Status") = 'Ordem Serviço' AND a."StInst" = 'True'
        THEN 'aguardando_retorno'::agenda_status
        WHEN TRIM(a."Status") = 'Ordem Serviço'
        THEN 'agendado'::agenda_status
        ELSE 'orcamento'::agenda_status
    END,
    -- tipo_evento
    CASE lower(TRIM(a."TipEvent"))
        WHEN 'licitação' THEN 'licitacao'::agenda_tipo_evento
        WHEN 'licitacao' THEN 'licitacao'::agenda_tipo_evento
        WHEN 'cortesia'  THEN 'cortesia'::agenda_tipo_evento
        WHEN 'recorrente' THEN 'recorrente'::agenda_tipo_evento
        ELSE 'particular'::agenda_tipo_evento
    END,
    -- tipo_retorno
    CASE
        WHEN TRIM(a."TipRet") ILIKE 'outra%' THEN 'outra_equipe'::agenda_tipo_retorno
        ELSE 'mesma_equipe'::agenda_tipo_retorno
    END,
    -- forma_pagamento
    CASE lower(TRIM(a."FormPag"))
        WHEN 'dinheiro'           THEN 'dinheiro'::forma_pagamento
        WHEN 'cheque'             THEN 'cheque'::forma_pagamento
        WHEN 'pix'                THEN 'pix'::forma_pagamento
        WHEN 'boleto'             THEN 'boleto'::forma_pagamento
        WHEN 'cartao'             THEN 'cartao'::forma_pagamento
        WHEN 'deposito bancario'  THEN 'transferencia'::forma_pagamento
        WHEN 'transferencia bancaria' THEN 'transferencia'::forma_pagamento
        ELSE NULL
    END,
    -- descricao_evento
    NULLIF(TRIM(a."DescEvento"), ''),
    -- data_evento
    NULLIF(TRIM(a."DtEvent"), '')::date,
    -- hora_evento: extrai parte time de timestamp bogus "1899-12-30 HH:MM:SS"
    CASE
        WHEN TRIM(a."HrEvent") ~ '^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$'
        THEN SUBSTRING(TRIM(a."HrEvent") FROM 12)::time
        ELSE NULL
    END,
    -- data_instalacao
    NULLIF(TRIM(a."datInst"), '')::date,
    -- hora_instalacao
    CASE
        WHEN TRIM(a."HrInst") ~ '^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$'
        THEN SUBSTRING(TRIM(a."HrInst") FROM 12)::time
        ELSE NULL
    END,
    -- data_retorno_prevista
    NULLIF(TRIM(a."DtRet"), '')::date,
    -- data_retorno_real
    NULLIF(TRIM(a."DtRealRet"), '')::date,
    -- numero_aprovacao
    CASE
        WHEN NULLIF(TRIM(a."NrAprov"), '') IN ('.', '..', '-') THEN NULL
        ELSE NULLIF(TRIM(a."NrAprov"), '')
    END,
    -- data_aprovacao: 2025-01-09 = placeholder do sistema legado → NULL
    CASE
        WHEN TRIM(a."DtAprov") LIKE '2025%' OR TRIM(a."DtAprov") LIKE '%0001%'
        THEN NULL
        ELSE NULLIF(TRIM(a."DtAprov"), '')::timestamptz
    END,
    -- data_cancelamento
    CASE
        WHEN TRIM(a."DtCanc") LIKE '2025%' OR TRIM(a."DtCanc") LIKE '%0001%'
        THEN NULL
        ELSE NULLIF(TRIM(a."DtCanc"), '')::timestamptz
    END,
    -- finalizado
    (a."StConc" = 'True'),
    -- observacoes
    NULLIF(TRIM(a."Obsagend"), ''),
    -- created_at: data de abertura da agenda
    COALESCE(NULLIF(TRIM(a."DtAg"), '')::date, CURRENT_DATE)::timestamptz
FROM _csv_agenda a
JOIN _agenda_map am ON CAST(CAST(TRIM(a."IdServ") AS float)::int AS int) = am.old_id
-- cliente pode ter sido deduplicado por documento duplicado
JOIN _cliente_doc_map cdm ON CAST(CAST(TRIM(a."IdClient") AS float)::int AS int) = cdm.old_id
LEFT JOIN _us_map usm ON CAST(TRIM(a."IdUs") AS int) = usm.old_id
-- motivo_cancelamento (MotCanc pode ser float ou vazio)
LEFT JOIN _motcanc_map mcm
    ON NULLIF(TRIM(a."MotCanc"), '') IS NOT NULL
   AND CAST(CAST(TRIM(a."MotCanc") AS float)::int AS int) = mcm.old_id
WHERE NULLIF(TRIM(a."IdServ"), '') IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 15. agenda_locais — cria um local principal por agenda com o endereco legado
-- ---------------------------------------------------------------------------

INSERT INTO agenda_locais (
    agenda_id, tipo, logradouro, numero, complemento, bairro, cidade,
    distancia_km, principal, ordem
)
SELECT
    am.new_id,
    'principal'::agenda_local_tipo,
    NULLIF(TRIM(a."EndEvento"), ''),
    NULLIF(TRIM(a."NrEnd"), ''),
    NULLIF(TRIM(a."ComplEnd"), ''),
    NULLIF(TRIM(a."Bairro"), ''),
    NULLIF(TRIM(a."Cidade"), ''),
    NULLIF(TRIM(a."DistEvent"), '')::numeric,
    TRUE,
    0
FROM _csv_agenda a
JOIN _agenda_map am ON CAST(CAST(TRIM(a."IdServ") AS float)::int AS int) = am.old_id
JOIN agenda ag ON ag.id = am.new_id
WHERE NULLIF(TRIM(a."IdServ"), '') IS NOT NULL
  AND (
      NULLIF(TRIM(a."EndEvento"), '') IS NOT NULL
      OR NULLIF(TRIM(a."Cidade"), '') IS NOT NULL
  );

-- ---------------------------------------------------------------------------
-- 16. agenda_local_contatos — contato principal e secundario do evento
-- ---------------------------------------------------------------------------

-- Contato principal
INSERT INTO agenda_local_contatos (
    agenda_local_id, nome, telefone_principal, principal
)
SELECT
    al.id,
    TRIM(a."ContEvent"),
    NULLIF(TRIM(a."TelContEvent"), ''),
    TRUE
FROM _csv_agenda a
JOIN _agenda_map am ON CAST(CAST(TRIM(a."IdServ") AS float)::int AS int) = am.old_id
JOIN agenda_locais al ON al.agenda_id = am.new_id AND al.principal = TRUE
WHERE NULLIF(TRIM(a."ContEvent"), '') IS NOT NULL;

-- Contato secundario
INSERT INTO agenda_local_contatos (
    agenda_local_id, nome, telefone_principal, principal
)
SELECT
    al.id,
    TRIM(a."Cont2Event"),
    NULLIF(TRIM(a."Tel2ContEvent"), ''),
    FALSE
FROM _csv_agenda a
JOIN _agenda_map am ON CAST(CAST(TRIM(a."IdServ") AS float)::int AS int) = am.old_id
JOIN agenda_locais al ON al.agenda_id = am.new_id AND al.principal = TRUE
WHERE NULLIF(TRIM(a."Cont2Event"), '') IS NOT NULL
  AND TRIM(a."Cont2Event") NOT IN ('.', '..', '-');

-- ---------------------------------------------------------------------------
-- 17. agenda_itens
-- ---------------------------------------------------------------------------

INSERT INTO agenda_itens (
    agenda_id, servico_locacao_id, numero_sequencial,
    descricao_complemento, quantidade, unidade,
    valor_unitario, valor_total, observacoes
)
SELECT
    am.new_id,
    lm.new_id,
    COALESCE(NULLIF(TRIM(i."NrServ"), '')::int, 0),
    NULLIF(TRIM(i."ComplItem"), ''),
    GREATEST(COALESCE(NULLIF(TRIM(i."Qdade"), '')::numeric, 1), 0.01),
    COALESCE(NULLIF(TRIM(i."Unid"), ''), 'DIARIA'),
    COALESCE(NULLIF(TRIM(i."VlrUnit"), '')::numeric, 0),
    COALESCE(NULLIF(TRIM(i."VlrTotal"), '')::numeric, 0),
    NULLIF(TRIM(i."ObsItem"), '')
FROM _csv_agenda_itens i
JOIN _agenda_map am ON CAST(CAST(TRIM(i."IdAgendaExt") AS float)::int AS int) = am.old_id
JOIN agenda ag ON ag.id = am.new_id
LEFT JOIN _loc_map lm
    ON NULLIF(TRIM(i."IdItemExt"), '') IS NOT NULL
   AND CAST(TRIM(i."IdItemExt") AS int) = lm.old_id
WHERE NULLIF(TRIM(i."IdAgendaExt"), '') IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 18. agenda_equipe — instalacao (TabEscala)
-- ---------------------------------------------------------------------------

INSERT INTO agenda_equipe (agenda_id, funcionario_id, papel)
SELECT
    am.new_id,
    fm.new_id,
    'instalacao'::agenda_equipe_papel
FROM _csv_escala e
JOIN _agenda_map am ON CAST(CAST(TRIM(e."IdServExt") AS float)::int AS int) = am.old_id
JOIN agenda ag ON ag.id = am.new_id
JOIN _func_map fm
    ON NULLIF(TRIM(e."IdFuncExt"), '') IS NOT NULL
   AND CAST(CAST(TRIM(e."IdFuncExt") AS float)::int AS int) = fm.old_id
WHERE NULLIF(TRIM(e."IdServExt"), '') IS NOT NULL
  AND NULLIF(TRIM(e."IdFuncExt"), '') IS NOT NULL
ON CONFLICT (agenda_id, funcionario_id, papel) DO NOTHING;

-- Retorno (TabEscalaRet)
INSERT INTO agenda_equipe (agenda_id, funcionario_id, papel)
SELECT
    am.new_id,
    fm.new_id,
    'retorno'::agenda_equipe_papel
FROM _csv_escala_ret e
JOIN _agenda_map am ON CAST(CAST(TRIM(e."IdServExt") AS float)::int AS int) = am.old_id
JOIN agenda ag ON ag.id = am.new_id
JOIN _func_map fm
    ON NULLIF(TRIM(e."IdFuncExt"), '') IS NOT NULL
   AND CAST(CAST(TRIM(e."IdFuncExt") AS float)::int AS int) = fm.old_id
WHERE NULLIF(TRIM(e."IdServExt"), '') IS NOT NULL
  AND NULLIF(TRIM(e."IdFuncExt"), '') IS NOT NULL
ON CONFLICT (agenda_id, funcionario_id, papel) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 19. agenda_veiculos — merge de saida + retorno por (evento, veiculo)
-- Ambos os arquivos têm KmSaida/KmRet vazios; usa UNION DISTINCT para
-- deduplicar pares (evento, veiculo) que aparecem nos dois arquivos.
-- ---------------------------------------------------------------------------

INSERT INTO agenda_veiculos (agenda_id, veiculo_id, km_saida, km_retorno)
SELECT DISTINCT ON (am.new_id, vm.new_id)
    am.new_id,
    vm.new_id,
    NULLIF(TRIM(s."KmSaida"), '')::numeric,
    NULLIF(TRIM(s."KmRet"),   '')::numeric
FROM (
    SELECT "IdEvento", "IdVeicExt", "KmSaida", "KmRet" FROM _csv_saida_veic
    UNION ALL
    SELECT "IdEvento", "IdVeicExt", "KmSaida", "KmRet" FROM _csv_saida_veic_ret
) s
JOIN _agenda_map am
    ON NULLIF(TRIM(s."IdEvento"), '') IS NOT NULL
   AND CAST(CAST(TRIM(s."IdEvento") AS float)::int AS int) = am.old_id
JOIN agenda ag ON ag.id = am.new_id
JOIN _veic_map vm
    ON NULLIF(TRIM(s."IdVeicExt"), '') IS NOT NULL
   AND CAST(CAST(TRIM(s."IdVeicExt") AS float)::int AS int) = vm.old_id
ORDER BY am.new_id, vm.new_id;

-- ---------------------------------------------------------------------------
-- 20. movimentacoes_estoque (TabEquipSaida → tipo saida)
-- ---------------------------------------------------------------------------

INSERT INTO movimentacoes_estoque (
    agenda_id, equipamento_id, tipo, quantidade, data_movimentacao
)
SELECT
    am.new_id,
    eqm.new_id,
    'saida'::movimentacao_tipo,
    -- quantidade como negativo (saída reduz estoque); QdaSaida vem como "8.0"
    -(COALESCE(NULLIF(TRIM(es."QdaSaida"), '')::float::int, 1)),
    -- data_movimentacao: usa data_instalacao da agenda quando disponível
    COALESCE(ag.data_instalacao::timestamptz, NOW())
FROM _csv_equip_saida es
JOIN _agenda_map am
    ON NULLIF(TRIM(es."IdAgendExt"), '') IS NOT NULL
   AND CAST(CAST(TRIM(es."IdAgendExt") AS float)::int AS int) = am.old_id
JOIN _equip_map eqm
    ON NULLIF(TRIM(es."IdItemExt"), '') IS NOT NULL
   AND CAST(CAST(TRIM(es."IdItemExt") AS float)::int AS int) = eqm.old_id
JOIN agenda ag ON ag.id = am.new_id
WHERE NULLIF(TRIM(es."IdAgendExt"), '') IS NOT NULL
  AND NULLIF(TRIM(es."IdItemExt"), '') IS NOT NULL;

-- ---------------------------------------------------------------------------
-- 21. contas_receber + recebimentos (TabPgto)
--
-- Estratégia: uma CR por agenda (agrupado) com valor = soma dos pagamentos;
-- cada linha do TabPgto vira um recebimento linkado à CR da agenda.
-- ---------------------------------------------------------------------------

-- CR agrupada por agenda
INSERT INTO contas_receber (
    tenant_id, agenda_id, cliente_id,
    competencia, data_emissao, data_vencimento,
    valor_original, valor_baixado, saldo, status
)
SELECT
    (SELECT tid FROM _t),
    am.new_id,
    ag.cliente_id,
    TO_CHAR(MIN(NULLIF(TRIM(p."DtRec"), '')::date), 'YYYY-MM'),
    MIN(NULLIF(TRIM(p."DtReg"), '')::date),
    MAX(NULLIF(TRIM(p."DtRec"), '')::date),
    SUM(COALESCE(NULLIF(TRIM(p."VlrRec"), '')::numeric, 0)),
    SUM(COALESCE(NULLIF(TRIM(p."VlrRec"), '')::numeric, 0)),
    0,
    'pago'::conta_status
FROM _csv_pgto p
JOIN _agenda_map am ON CAST(CAST(TRIM(p."IdAgendExt") AS float)::int AS int) = am.old_id
JOIN agenda ag ON ag.id = am.new_id
WHERE NULLIF(TRIM(p."IdAgendExt"), '') IS NOT NULL
GROUP BY am.new_id, ag.cliente_id;

-- Recebimentos individuais
INSERT INTO recebimentos (
    conta_receber_id, usuario_registro_id,
    data_recebimento, valor_recebido,
    forma_pagamento, tipo_documento, numero_documento
)
SELECT
    cr.id,
    usm.new_id,
    COALESCE(NULLIF(TRIM(p."DtRec"), '')::date, CURRENT_DATE),
    COALESCE(NULLIF(TRIM(p."VlrRec"), '')::numeric, 0),
    CASE lower(TRIM(p."FormRec"))
        WHEN 'dinheiro'               THEN 'dinheiro'::forma_pagamento
        WHEN 'cheque'                 THEN 'cheque'::forma_pagamento
        WHEN 'pix'                    THEN 'pix'::forma_pagamento
        WHEN 'boleto'                 THEN 'boleto'::forma_pagamento
        WHEN 'cartao'                 THEN 'cartao'::forma_pagamento
        WHEN 'transferencia bancaria' THEN 'transferencia'::forma_pagamento
        WHEN 'deposito bancario'      THEN 'transferencia'::forma_pagamento
        ELSE 'dinheiro'::forma_pagamento
    END,
    CASE lower(TRIM(p."DocRec"))
        WHEN 'nota fiscal' THEN 'nota_fiscal'::tipo_documento_fiscal
        WHEN 'cupom'       THEN 'cupom'::tipo_documento_fiscal
        WHEN 'recibo'      THEN 'recibo'::tipo_documento_fiscal
        ELSE               NULL
    END,
    CASE
        WHEN COALESCE(NULLIF(TRIM(p."NrDocRec"), ''), '0') = '0' THEN NULL
        ELSE NULLIF(TRIM(p."NrDocRec"), '')
    END
FROM _csv_pgto p
JOIN _agenda_map am ON CAST(CAST(TRIM(p."IdAgendExt") AS float)::int AS int) = am.old_id
-- Liga ao CR da agenda correspondente
JOIN contas_receber cr
    ON cr.agenda_id = am.new_id
   AND cr.tenant_id = (SELECT tid FROM _t)
LEFT JOIN _us_map usm ON CAST(TRIM(p."IdUs") AS int) = usm.old_id
WHERE NULLIF(TRIM(p."IdAgendExt"), '') IS NOT NULL
  AND COALESCE(NULLIF(TRIM(p."VlrRec"), '')::numeric, 0) <> 0;

-- ---------------------------------------------------------------------------
-- 22. Atualiza valor_total da agenda com base nos itens importados
-- ---------------------------------------------------------------------------

UPDATE agenda ag
SET
    valor_total   = COALESCE(totais.total, 0),
    valor_liquido = COALESCE(totais.total, 0)
FROM (
    SELECT agenda_id, SUM(valor_total) AS total
    FROM agenda_itens
    GROUP BY agenda_id
) totais
WHERE ag.id = totais.agenda_id
  AND ag.tenant_id = (SELECT tid FROM _t);

-- ---------------------------------------------------------------------------
-- 23. Admin do tenant: usuario "admin" com role sistema admin
-- ---------------------------------------------------------------------------

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM usuarios u
CROSS JOIN roles r
WHERE u.tenant_id = (SELECT tid FROM _t)
  AND u.papel = 'admin'::usuario_papel
  AND r.codigo = 'admin'
  AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------------
-- 24. Admin principal do tenant: credenciais definitivas
--
-- O CSV importa 6 usuarios com logins numericos (IdUsuario = 1..6) e
-- emails placeholder (@radelgo.local). Esta secao:
--   1. Desativa o "Administrador" generico (login='1') do CSV
--   2. Promove Vilson Emilio a admin com slug, email e senha reais
--
-- Credenciais de acesso iniciais:
--   Organizacao : radelgo
--   E-mail      : vilson.emilio@radelgo.local
--   Senha       : 12345678  (trocar no primeiro acesso)
-- ---------------------------------------------------------------------------

-- Desativa o usuario "Administrador" generico importado pelo CSV
UPDATE usuarios SET ativo = FALSE, papel = 'comercial'
WHERE tenant_id = (SELECT tid FROM _t)
  AND login = '1';

-- Promove Vilson Emilio: troca login numerico por slug, seta email e senha reais
UPDATE usuarios SET
    login      = 'vilson.emilio',
    email      = 'vilson.emilio@radelgo.local',
    nome       = 'Vilson Emilio',
    senha_hash = crypt('12345678', gen_salt('bf', 12)),
    papel      = 'admin',
    ativo      = TRUE
WHERE tenant_id = (SELECT tid FROM _t)
  AND LOWER(nome) LIKE '%vilson%emilio%';

-- Garante role admin (idempotente mesmo se papel foi elevado acima)
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM usuarios u, roles r
WHERE u.tenant_id = (SELECT tid FROM _t)
  AND u.login     = 'vilson.emilio'
  AND r.codigo    = 'admin'
  AND r.tenant_id IS NULL
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------------
-- 25. Resumo da importacao
-- ---------------------------------------------------------------------------

DO $$
DECLARE
    v_tid uuid;
BEGIN
    SELECT tid INTO v_tid FROM _t LIMIT 1;

    RAISE NOTICE '========================================================';
    RAISE NOTICE 'Tenant: radelgo  (id: %)', v_tid;
    RAISE NOTICE 'classificacoes_servico : %', (SELECT COUNT(*) FROM classificacoes_servico WHERE tenant_id = v_tid);
    RAISE NOTICE 'servicos_locacao       : %', (SELECT COUNT(*) FROM servicos_locacao       WHERE tenant_id = v_tid);
    RAISE NOTICE 'funcionarios           : %', (SELECT COUNT(*) FROM funcionarios           WHERE tenant_id = v_tid);
    RAISE NOTICE 'usuarios               : %', (SELECT COUNT(*) FROM usuarios               WHERE tenant_id = v_tid);
    RAISE NOTICE 'motivos_cancelamento   : %', (SELECT COUNT(*) FROM motivos_cancelamento   WHERE tenant_id = v_tid);
    RAISE NOTICE 'clientes               : %', (SELECT COUNT(*) FROM clientes               WHERE tenant_id = v_tid);
    RAISE NOTICE 'veiculos               : %', (SELECT COUNT(*) FROM veiculos               WHERE tenant_id = v_tid);
    RAISE NOTICE 'equipamentos           : %', (SELECT COUNT(*) FROM equipamentos           WHERE tenant_id = v_tid);
    RAISE NOTICE 'kit_composicao         : %', (SELECT COUNT(*) FROM kit_composicao kc JOIN servicos_locacao sl ON sl.id = kc.servico_locacao_id WHERE sl.tenant_id = v_tid);
    RAISE NOTICE 'agenda                 : %', (SELECT COUNT(*) FROM agenda                 WHERE tenant_id = v_tid);
    RAISE NOTICE 'agenda_itens           : %', (SELECT COUNT(*) FROM agenda_itens ai JOIN agenda ag ON ag.id = ai.agenda_id WHERE ag.tenant_id = v_tid);
    RAISE NOTICE 'agenda_equipe          : %', (SELECT COUNT(*) FROM agenda_equipe ae JOIN agenda ag ON ag.id = ae.agenda_id WHERE ag.tenant_id = v_tid);
    RAISE NOTICE 'agenda_veiculos        : %', (SELECT COUNT(*) FROM agenda_veiculos av JOIN agenda ag ON ag.id = av.agenda_id WHERE ag.tenant_id = v_tid);
    RAISE NOTICE 'movimentacoes_estoque  : %', (SELECT COUNT(*) FROM movimentacoes_estoque me JOIN agenda ag ON ag.id = me.agenda_id WHERE ag.tenant_id = v_tid);
    RAISE NOTICE 'contas_receber         : %', (SELECT COUNT(*) FROM contas_receber         WHERE tenant_id = v_tid);
    RAISE NOTICE 'recebimentos           : %', (SELECT COUNT(*) FROM recebimentos r JOIN contas_receber cr ON cr.id = r.conta_receber_id WHERE cr.tenant_id = v_tid);
    RAISE NOTICE '========================================================';
END $$;

COMMIT;
