# Especificacao Funcional - ERP de Locacoes e Leasing de Projetos

## 1. Contexto

Este documento descreve o comportamento esperado do ERP de locacoes e leasing de projetos, cobrindo operacao comercial, financeira, fiscal e operacional. A proposta e transformar o processo em um fluxo unico, rastreavel e auditavel, desde o orçamento ate o encerramento.

As imagens de referencia no repositorio indicam um sistema com alto volume de dados, navegação por abas e foco em operacao de campo, o que foi considerado nesta especificacao.

## 2. Objetivo do Produto

O sistema deve permitir:

- cadastrar clientes, funcionarios, frota, inventario e usuarios;
- criar orcamentos e aprovar propostas;
- agendar servicos, executar instalacoes e registrar retiradas;
- controlar valores a receber, recebimentos e saldo;
- apoiar emissao fiscal e rastreabilidade documental;
- acompanhar ponto, localizacao e produtividade de equipes;
- manter historico de operacao e auditoria de alteracoes.

## 3. Perfis de Usuario

- Administrador
- Comercial
- Operacao
- Financeiro
- Fiscal
- RH / DP
- Almoxarifado
- Supervisor de campo
- Diretoria

Cada perfil deve enxergar apenas as funcoes necessarias para sua atividade.

## 4. Escopo Funcional

### 4.1 Comercial

- cadastro e consulta de clientes;
- criacao de orcamentos;
- composicao de itens, kits, taxa de deslocamento e instalacao;
- definicao de forma de pagamento e condicoes comerciais;
- aprovacao ou rejeicao da proposta;
- conversao do orcamento em ordem de servico.

### 4.2 Operacao

- agenda de servicos;
- reserva de equipe, frota e inventario;
- instalacao;
- acompanhamento de status;
- retirada de itens;
- conferencia de retorno;
- encerramento da ordem de servico.

### 4.3 Financeiro

- contas a receber;
- registro de parcelas e adiantamentos;
- baixas manuais e automáticas;
- controle de saldo;
- recebimentos parciais e totais;
- relatórios de inadimplencia e faturamento.

### 4.4 Fiscal

- controle de documentos fiscais;
- associacao de nota ao servico;
- validacao de dados obrigatorios;
- cancelamento e reemissao com trilha de auditoria;
- controle por tipo de operacao.

### 4.5 Recursos Humanos e Campo

- cadastro de funcionarios;
- escalas e funcoes;
- ponto;
- localizacao em campo;
- produtividade por equipe;
- registro de responsavel por etapa.

### 4.6 Frota e Inventario

- cadastro de veiculos;
- controle de km de saida e retorno;
- manutencao e disponibilidade;
- cadastro de equipamentos e kits;
- reservas, separacao e retorno;
- baixa por avaria, perda ou descarte.

## 5. Fluxos Principais

### 5.1 Fluxo de Orçamento

1. usuario cria orcamento;
2. vincula cliente e local;
3. adiciona itens, prazos e valores;
4. grava observacoes e anexos;
5. envia para aprovacao.

Regras:

- orcamento deve possuir cliente;
- valores devem ser calculados antes do envio;
- a proposta pode expirar por prazo de validade.

### 5.2 Fluxo de Aprovacao

1. aprovador analisa proposta;
2. aprova, reprova ou devolve para ajuste;
3. sistema registra data, usuario e observacao;
4. aprovacao libera o agendamento.

Regras:

- servicos acima de limite podem exigir aprovacao adicional;
- status de aprovacao deve ser historico e rastreavel.

### 5.3 Fluxo de Agendamento

1. ordem de servico e criada;
2. equipe, frota e inventario sao reservados;
3. datas de instalacao e retorno sao definidas;
4. servico entra em fila operacional.

Regras:

- nao permitir conflito de reserva de recurso;
- permitir reagendamento sem perder historico;
- manter vinculo entre OS e orçamento original.

### 5.4 Fluxo de Execucao

1. equipe confirma saida;
2. sistema registra deslocamento e chegada;
3. instalacao e executada;
4. fotos, observacoes e ocorrencias sao adicionadas;
5. status do servico e atualizado.

Regras:

- confirmar equipe responsavel;
- impedir fechamento sem registro minimo de execucao;
- guardar evidencias quando aplicavel.

### 5.5 Fluxo de Retorno e Encerramento

1. itens sao conferidos;
2. veiculo retorna com km final;
3. pendencias e avarias sao registradas;
4. servico e encerrado.

Regras:

- nao encerrar sem confirmar retorno;
- qualquer divergencia deve gerar ocorrencia;
- encerramento pode bloquear financeiro se houver pendencia critica.

### 5.6 Fluxo Financeiro

1. servico concluido gera base de cobranca;
2. financeiro emite titulo;
3. pagamentos sao registrados;
4. saldo e atualizado;
5. inadimplencia e monitorada.

## 6. Requisitos por Módulo

### 6.1 Clientes

- cadastrar pessoa fisica ou juridica;
- guardar contatos, enderecos e observacoes;
- manter historico de atendimento.

### 6.2 Orçamentos

- permitir itens avulsos e kits;
- calcular desconto e adicional;
- exibir valor total e prazo de validade;
- registrar responsavel pela negociacao.

### 6.3 Ordens de Servico

- exibir numero unico;
- associar cliente, evento e responsavel;
- guardar status;
- vincular equipe, frota e inventario;
- manter historico de status.

### 6.4 Financeiro

- registrar valor previsto, faturado e recebido;
- mostrar saldo aberto;
- registrar forma de recebimento;
- suportar baixa parcial.

### 6.5 Fiscal

- emitir documento com base no servico;
- exigir campos obrigatorios antes da emissao;
- manter trilha de alteracoes e cancelamentos.

### 6.6 Funcionarios

- relacionar colaborador a funcoes;
- registrar alocacao por servico;
- controlar ponto e deslocamento.

### 6.7 Frota

- mostrar disponibilidade;
- controlar manutencao;
- registrar km inicial e final;
- associar motorista e responsavel.

### 6.8 Inventario

- controlar item, quantidade e status;
- reservar materiais por ordem;
- registrar retirada e retorno;
- tratar avarias e baixas.

## 7. Estados do Sistema

### 7.1 Status do Orcamento

- rascunho
- em analise
- aprovado
- recusado
- expirado
- convertido

### 7.2 Status da Ordem de Servico

- criado
- agendado
- separado
- em execucao
- aguardando retorno
- finalizado
- cancelado

### 7.3 Status do Inventario

- disponivel
- reservado
- em campo
- retornado
- avariado
- perdido
- manutencao

### 7.4 Status da Frota

- disponivel
- reservado
- em rota
- em manutencao
- indisponivel

## 8. Validações Funcionais

- nao permitir servico sem cliente;
- nao permitir agenda com conflito de recurso;
- nao permitir saida de item indisponivel;
- nao permitir encerramento sem retorno;
- nao permitir faturamento sem origem valida;
- nao permitir mudanca fiscal sem permissao;
- nao permitir finalizacao com pendencia critica aberta.

## 9. Auditoria

Deve ser registrado:

- usuario;
- data e hora;
- operacao realizada;
- valor anterior;
- novo valor;
- motivo;
- origem da alteracao.

Eventos auditaveis:

- criacao;
- edicao;
- aprovacao;
- cancelamento;
- baixa;
- emissao fiscal;
- encerramento;
- alteracao de status.

## 10. Relatorios Funcionais

- servicos por status;
- servicos por cliente;
- servicos por periodo;
- faturamento por cliente;
- saldo em aberto;
- inventario em uso;
- frota por periodo;
- produtividade por colaborador;
- atraso de retorno;
- aprovacoes pendentes.

## 11. Requisitos de Experiencia

- telas com poucos cliques para a tarefa principal;
- filtros claros por cliente, data e status;
- navegação por abas;
- validacao imediata de campos;
- uso em desktop com grande volume de dados;
- apoio em mobile para consulta e campo.

## 12. Critérios de Aceite

Um fluxo esta aprovado quando:

- salva o historico completo;
- respeita as permissoes;
- atualiza os modulos relacionados;
- permite consulta posterior;
- nao perde rastreabilidade;
- gera resultado operacional consistente.

## 13. Questões em Aberto

- o sistema tera contratos recorrentes ou apenas projetos avulsos?;
- notas fiscais serao emitidas internamente ou por integracao?;
- a localizacao sera em tempo real ou apenas por check-in?;
- o ponto integrara folha de pagamento?;
- o portal do cliente faz parte do MVP ou da fase avancada?
