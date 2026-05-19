# Mapa de Telas - ERP de Locacoes e Leasing de Projetos

## 1. Objetivo

Organizar a navegacao do ERP por areas funcionais, mantendo o fluxo operacional rapido e coerente com o dominio de locacoes, leasing e servicos.

## 2. Estrutura Principal de Navegacao

### 2.1 Inicio / Dashboard

- visao geral de servicos do dia;
- alertas de pendencia;
- indicadores financeiros;
- avisos de retorno, atraso e aprovacao.

### 2.2 Comercial

- clientes;
- contatos;
- orcamentos;
- aprovacoes;
- conversao em ordem de servico.

### 2.3 Operacao

- agenda de servicos;
- ordens de servico;
- instalacoes;
- retiradas;
- ocorrencias operacionais.

### 2.4 Financeiro

- contas a receber;
- recebimentos;
- baixas;
- saldos;
- inadimplencia.

### 2.5 Fiscal

- documentos fiscais;
- emissao;
- cancelamento;
- reemissao;
- historico fiscal.

### 2.6 Recursos

- funcionarios;
- escalas;
- ponto;
- localizacao;
- frota;
- inventario.

### 2.7 Administracao

- usuarios;
- perfis;
- permissoes;
- parametros;
- logs de auditoria.

## 3. Telas por Modulo

### 3.1 Dashboard

Elementos:

- cards de status;
- grafico de faturamento;
- lista de servicos atrasados;
- servicos aguardando retorno;
- pendencias de aprovacao.

### 3.2 Clientes

Componentes:

- busca;
- filtro por nome e documento;
- cadastro principal;
- contatos;
- enderecos;
- historico de atendimento.

### 3.3 Orcamentos

Componentes:

- cabecalho com cliente e responsavel;
- itens do orcamento;
- resumo financeiro;
- status;
- aprovacao;
- anexos.

### 3.4 Ordens de Servico

Componentes:

- dados gerais;
- local;
- datas;
- equipe;
- frota;
- inventario;
- status historico;
- observacoes;
- anexos.

### 3.5 Agenda

Componentes:

- lista por data;
- filtro por status;
- drag and drop opcional;
- reserva de recursos;
- reagendamento.

### 3.6 Financeiro

Componentes:

- titulos em aberto;
- pagamentos;
- saldo por cliente;
- filtros por competencia;
- exportacao.

### 3.7 Fiscal

Componentes:

- documentos emitidos;
- pendentes;
- cancelados;
- chave de acesso;
- arquivos XML e PDF.

### 3.8 Funcionarios

Componentes:

- cadastro;
- funcoes;
- escalas;
- produtividade;
- historico de servicos.

### 3.9 Frota

Componentes:

- cadastro do veiculo;
- disponibilidade;
- km;
- manutencao;
- historico de uso.

### 3.10 Inventario

Componentes:

- itens;
- kits;
- reservas;
- saidas;
- retornos;
- avarias.

### 3.11 Ponto e Localizacao

Componentes:

- marcacoes de ponto;
- mapa ou coordenadas;
- jornada;
- atrasos;
- check-in e check-out.

### 3.12 Administracao

Componentes:

- usuarios;
- perfis;
- permissoes;
- logs;
- configuracoes;
- sequencias numericas.

## 4. Fluxo de Navegacao Recomendado

1. dashboard;
2. cliente ou orcamento;
3. aprovacao;
4. ordem de servico;
5. agenda e separacao;
6. execucao em campo;
7. retirada e retorno;
8. financeiro e fiscal;
9. auditoria e relatorios.

## 5. Regras de UX

- manter identidade visual unica por modulo;
- destacar status com cor e texto;
- reduzir excesso de cliques;
- exibir dados mais importantes no topo;
- manter historico acessivel sem abrir varias telas;
- permitir busca rapida em qualquer lista.

## 6. Telas Críticas

As telas que mais exigem clareza e estabilidade sao:

- orcamento;
- ordem de servico;
- agenda;
- controle financeiro;
- retirada e retorno;
- documento fiscal;
- dashboard operacional.

## 7. Observacao

Este mapa serve como base para prototipacao, desenvolvimento e validação com usuarios finais.
