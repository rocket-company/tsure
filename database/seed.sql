-- =============================================================================
-- Seed minimo para dev. Roda APOS schema.sql.
-- O usuario admin nao e criado aqui: a aplicacao Go faz bootstrap do admin
-- no startup (apps/web/internal/auth) usando bcrypt em tempo de execucao.
-- Defina ADMIN_PASSWORD no ambiente; padrao "tsure-admin" se nao informado.
-- =============================================================================

INSERT INTO classificacoes_servico (codigo, descricao, ordem) VALUES
    ('palco',         'Palco',              1),
    ('tenda',         'Tenda',              2),
    ('sonorizacao',   'Sonorizacao',        3),
    ('iluminacao',    'Iluminacao',         4),
    ('grupo_gerador', 'Grupo Gerador',      5),
    ('mesas_cadeiras','Mesas e Cadeiras',   6),
    ('banheiros',     'Banheiros',          7),
    ('climatizador',  'Climatizador',       8),
    ('caixa_termica', 'Caixa Termica',      9),
    ('estrutura',     'Estrutura',         10),
    ('extras',        'Extras',            11)
ON CONFLICT (codigo) DO NOTHING;

INSERT INTO motivos_cancelamento (codigo, descricao) VALUES
    ('cliente_desistiu',  'Cliente desistiu'),
    ('chuva',             'Condicoes climaticas'),
    ('falta_pagamento',   'Falta de pagamento'),
    ('reagendado',        'Reagendado para outra data'),
    ('forca_maior',       'Forca maior')
ON CONFLICT (codigo) DO NOTHING;
