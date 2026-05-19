# Modelo de Banco de Dados - ERP de Locacoes e Leasing de Projetos

## 1. Premissas

O modelo deve priorizar:

- rastreabilidade;
- separacao por dominios;
- historico de alteracoes;
- controle de recursos;
- suporte a operacao e financeiro no mesmo fluxo.

O desenho abaixo assume um banco relacional, com possibilidade de evolucao para visoes, auditoria dedicada e armazenamento de anexos em repositorio externo.

## 2. Diretrizes de Modelagem

- usar chaves primarias tecnicas;
- manter chaves naturais apenas como restricao unica quando necessario;
- registrar `created_at`, `updated_at`, `deleted_at` quando aplicavel;
- manter `created_by`, `updated_by` para auditoria;
- evitar sobrescrever historico importante;
- usar tabelas de historico para status e alteracoes sensiveis;
- indexar campos de busca frequente.

## 3. Entidades de Alto Nivel

- clientes
- contatos
- enderecos
- orcamentos
- orcamento_itens
- aprovacoes
- ordens_servico
- ordem_servico_status_historico
- agendamentos
- instalacoes
- retiradas
- ocorrencias_operacionais
- contas_receber
- recebimentos
- documentos_fiscais
- funcionarios
- escalas
- pontos
- localizacoes
- veiculos
- manutencoes_veiculos
- equipamentos
- kits
- itens_kit
- movimentacoes_estoque
- usuarios
- perfis
- permissoes
- logs_auditoria
- anexos

## 4. Dicionario de Dados Sugerido

### 4.1 clientes

- `id`
- `razao_social`
- `nome_fantasia`
- `tipo_pessoa`
- `documento`
- `inscricao_estadual`
- `telefone`
- `email`
- `status`
- `observacoes`
- `created_at`
- `updated_at`
- `deleted_at`

Indices:

- unique em `documento`
- index em `razao_social`

### 4.2 contatos

- `id`
- `cliente_id`
- `nome`
- `cargo`
- `telefone`
- `celular`
- `email`
- `principal`
- `observacoes`

Relacionamento:

- muitos contatos para um cliente

### 4.3 enderecos

- `id`
- `cliente_id`
- `tipo`
- `logradouro`
- `numero`
- `complemento`
- `bairro`
- `cidade`
- `estado`
- `cep`
- `referencia`
- `principal`

### 4.4 orcamentos

- `id`
- `cliente_id`
- `responsavel_comercial_id`
- `numero`
- `data_orcamento`
- `validade`
- `status`
- `tipo_evento`
- `forma_pagamento`
- `valor_total`
- `desconto_total`
- `observacoes`
- `aprovado_por`
- `aprovado_em`

Indices:

- unique em `numero`
- index em `cliente_id`
- index em `status`

### 4.5 orcamento_itens

- `id`
- `orcamento_id`
- `tipo_item`
- `item_id`
- `descricao`
- `quantidade`
- `valor_unitario`
- `valor_total`
- `observacoes`

Observacao:

- pode referenciar item de catalogo ou texto livre.

### 4.6 aprovacoes

- `id`
- `orcamento_id`
- `usuario_id`
- `status`
- `data_aprovacao`
- `motivo_reprovacao`
- `numero_aprovacao`
- `observacoes`

### 4.7 ordens_servico

- `id`
- `orcamento_id`
- `cliente_id`
- `numero`
- `tipo_evento`
- `status`
- `data_agendada`
- `hora_evento`
- `data_instalacao`
- `hora_instalacao`
- `data_prevista_retorno`
- `responsavel_operacional_id`
- `responsavel_evento`
- `telefone_contato`
- `endereco_evento_id`
- `forma_pagamento`
- `valor_total`
- `valor_receber`
- `valor_recebido`
- `saldo`
- `observacoes`
- `finalizado_em`

Indices:

- unique em `numero`
- index em `cliente_id`
- index em `status`
- index em `data_agendada`

### 4.8 ordem_servico_status_historico

- `id`
- `ordem_servico_id`
- `status_anterior`
- `status_novo`
- `usuario_id`
- `data_movimento`
- `motivo`

### 4.9 agendamentos

- `id`
- `ordem_servico_id`
- `data_inicio`
- `data_fim`
- `status`
- `observacoes`

### 4.10 instalacoes

- `id`
- `ordem_servico_id`
- `usuario_responsavel_id`
- `data_inicio`
- `data_fim`
- `localizacao_id`
- `observacoes`
- `anexos_habilitados`

### 4.11 retiradas

- `id`
- `ordem_servico_id`
- `usuario_responsavel_id`
- `tipo_retirada`
- `data_retirada`
- `data_retorno`
- `km_saida`
- `km_retorno`
- `observacoes`

### 4.12 ocorrencias_operacionais

- `id`
- `ordem_servico_id`
- `tipo_ocorrencia`
- `gravidade`
- `descricao`
- `usuario_id`
- `data_ocorrencia`
- `resolvida`

### 4.13 contas_receber

- `id`
- `ordem_servico_id`
- `cliente_id`
- `numero_titulo`
- `competencia`
- `data_emissao`
- `data_vencimento`
- `valor_original`
- `valor_baixado`
- `saldo`
- `status`
- `forma_recebimento`
- `observacoes`

Indices:

- index em `cliente_id`
- index em `data_vencimento`
- index em `status`

### 4.14 recebimentos

- `id`
- `conta_receber_id`
- `data_recebimento`
- `valor_recebido`
- `meio_pagamento`
- `referencia`
- `usuario_id`
- `observacoes`

### 4.15 documentos_fiscais

- `id`
- `ordem_servico_id`
- `cliente_id`
- `numero`
- `serie`
- `tipo_documento`
- `status`
- `data_emissao`
- `valor_total`
- `base_calculo`
- `valor_imposto`
- `chave_acesso`
- `xml_path`
- `pdf_path`
- `observacoes`

### 4.16 funcionarios

- `id`
- `nome`
- `documento`
- `cargo`
- `funcao`
- `telefone`
- `email`
- `status`
- `observacoes`

### 4.17 escalas

- `id`
- `funcionario_id`
- `ordem_servico_id`
- `data_inicio`
- `data_fim`
- `funcao_na_os`
- `status`

### 4.18 pontos

- `id`
- `funcionario_id`
- `ordem_servico_id`
- `data_ponto`
- `hora_entrada`
- `hora_saida`
- `total_horas`
- `tipo_registro`
- `observacoes`

### 4.19 localizacoes

- `id`
- `funcionario_id`
- `ordem_servico_id`
- `latitude`
- `longitude`
- `endereco_texto`
- `data_localizacao`
- `origem`

### 4.20 veiculos

- `id`
- `placa`
- `modelo`
- `marca`
- `ano`
- `cor`
- `km_atual`
- `status`
- `motorista_padrao_id`
- `observacoes`

Indices:

- unique em `placa`

### 4.21 manutencoes_veiculos

- `id`
- `veiculo_id`
- `data_manutencao`
- `tipo`
- `descricao`
- `km`
- `custo`
- `status`

### 4.22 equipamentos

- `id`
- `codigo_patrimonio`
- `nome`
- `descricao`
- `categoria`
- `quantidade_total`
- `quantidade_disponivel`
- `status`
- `observacoes`

### 4.23 kits

- `id`
- `nome`
- `descricao`
- `status`
- `observacoes`

### 4.24 itens_kit

- `id`
- `kit_id`
- `equipamento_id`
- `quantidade`

### 4.25 movimentacoes_estoque

- `id`
- `equipamento_id`
- `kit_id`
- `ordem_servico_id`
- `tipo_movimentacao`
- `quantidade`
- `data_movimentacao`
- `usuario_id`
- `origem`
- `destino`
- `observacoes`

### 4.26 usuarios

- `id`
- `nome`
- `login`
- `email`
- `senha_hash`
- `status`
- `perfil_id`
- `ultimo_acesso`

### 4.27 perfis

- `id`
- `nome`
- `descricao`
- `status`

### 4.28 permissoes

- `id`
- `perfil_id`
- `codigo`
- `descricao`
- `ativo`

### 4.29 logs_auditoria

- `id`
- `usuario_id`
- `entidade`
- `registro_id`
- `acao`
- `valor_anterior`
- `valor_novo`
- `motivo`
- `data_evento`
- `ip_origem`

### 4.30 anexos

- `id`
- `entidade`
- `registro_id`
- `nome_arquivo`
- `caminho_arquivo`
- `tipo_mime`
- `tamanho`
- `usuario_id`
- `data_upload`

## 5. Relacionamentos Principais

- cliente 1:N contatos
- cliente 1:N enderecos
- cliente 1:N orcamentos
- orcamento 1:N orcamento_itens
- orcamento 1:1 aprovacao
- orcamento 1:1 ordem_servico
- ordem_servico 1:N agendamentos
- ordem_servico 1:N instalacoes
- ordem_servico 1:N retiradas
- ordem_servico 1:N ocorrencias_operacionais
- ordem_servico 1:N contas_receber
- contas_receber 1:N recebimentos
- ordem_servico 1:N documentos_fiscais
- funcionario 1:N escalas
- funcionario 1:N pontos
- funcionario 1:N localizacoes
- veiculo 1:N manutencoes_veiculos
- kit 1:N itens_kit
- equipamento 1:N itens_kit
- equipamento 1:N movimentacoes_estoque

## 6. Restrições de Integridade

- `orcamentos.numero` deve ser unico;
- `ordens_servico.numero` deve ser unico;
- `veiculos.placa` deve ser unica;
- `clientes.documento` deve ser unico por tipo de pessoa;
- `contas_receber.saldo` nao pode ser negativo sem regra explicita;
- `quantidade_disponivel` nao pode ficar abaixo de zero;
- `data_retorno` nao pode ser menor que `data_instalacao`;
- `status` deve seguir lista valida por entidade;
- exclusao fisica deve ser evitada em registros centrais.

## 7. Indexes Recomendados

- cliente / documento;
- status de ordem de servico;
- data de agendamento;
- data de vencimento;
- placa do veiculo;
- funcionario / data;
- inventario / status;
- logs por entidade e registro;
- ordens por responsavel.

## 8. Historico e Soft Delete

Para entidades criticas, usar:

- `deleted_at` para exclusao logica;
- tabelas de historico para status;
- tabelas de eventos para alteracoes sensiveis.

Entidades que merecem historico reforcado:

- orcamentos;
- ordens de servico;
- financeiro;
- documentos fiscais;
- movimentacoes de estoque;
- pontos;
- localizacoes;
- logs de aprovacao.

## 9. Sugestao de Separaçao por Schemas

Se o banco suportar schemas, uma organizacao possivel e:

- `comercial`
- `operacao`
- `financeiro`
- `fiscal`
- `rh`
- `estoque`
- `seguranca`
- `auditoria`

Isso melhora leitura, manutencao e evolucao.

## 10. Dados Sensiveis

Campos sensiveis ou de atencao:

- documento do cliente;
- senha do usuario;
- chaves de acesso fiscal;
- coordenadas de localizacao;
- historico de ponto;
- valores financeiros;
- observacoes operacionais com dados pessoais.

## 11. Consideracoes Finais

O modelo precisa favorecer a operacao real. Em vez de enxergar apenas cadastros, o banco deve refletir o ciclo completo: proposta, aprovacao, agenda, execucao, retorno, faturamento e auditoria.
