# API Spec - ERP de Locacoes e Leasing de Projetos

## 1. Objetivo

Definir uma API para suportar o ERP com foco em operacao, financeiro, fiscal, recursos e auditoria.

## 2. Convencoes Gerais

- REST com JSON.
- autenticacao por token.
- respostas padronizadas.
- paginacao em listagens.
- filtros por query string.
- ids tecnicos nas entidades.

### 2.1 Formato de Resposta

```json
{
  "data": {},
  "message": "ok",
  "errors": []
}
```

### 2.2 Erros Padronizados

- `400` validacao;
- `401` nao autenticado;
- `403` sem permissao;
- `404` nao encontrado;
- `409` conflito;
- `422` regra de negocio;
- `500` erro interno.

## 3. Modulos de API

### 3.1 Autenticacao

- `POST /auth/login`
- `POST /auth/logout`
- `GET /auth/me`

### 3.2 Clientes

- `GET /clientes`
- `POST /clientes`
- `GET /clientes/{id}`
- `PUT /clientes/{id}`
- `DELETE /clientes/{id}`

Sub-recursos:

- `/clientes/{id}/contatos`
- `/clientes/{id}/enderecos`
- `/clientes/{id}/historico`

### 3.3 Orcamentos

- `GET /orcamentos`
- `POST /orcamentos`
- `GET /orcamentos/{id}`
- `PUT /orcamentos/{id}`
- `POST /orcamentos/{id}/aprovar`
- `POST /orcamentos/{id}/reprovar`
- `POST /orcamentos/{id}/converter-os`

### 3.4 Ordens de Servico

- `GET /ordens-servico`
- `POST /ordens-servico`
- `GET /ordens-servico/{id}`
- `PUT /ordens-servico/{id}`
- `POST /ordens-servico/{id}/agendar`
- `POST /ordens-servico/{id}/instalar`
- `POST /ordens-servico/{id}/retirar`
- `POST /ordens-servico/{id}/finalizar`
- `POST /ordens-servico/{id}/cancelar`

### 3.5 Agenda

- `GET /agenda`
- `POST /agenda`
- `PUT /agenda/{id}`
- `POST /agenda/{id}/reagendar`

### 3.6 Financeiro

- `GET /financeiro/contas-receber`
- `POST /financeiro/contas-receber`
- `GET /financeiro/contas-receber/{id}`
- `POST /financeiro/contas-receber/{id}/baixar`
- `POST /financeiro/contas-receber/{id}/negociar`
- `GET /financeiro/recebimentos`

### 3.7 Fiscal

- `GET /fiscal/documentos`
- `POST /fiscal/documentos`
- `GET /fiscal/documentos/{id}`
- `POST /fiscal/documentos/{id}/cancelar`
- `POST /fiscal/documentos/{id}/reemissao`

### 3.8 Funcionarios

- `GET /funcionarios`
- `POST /funcionarios`
- `GET /funcionarios/{id}`
- `PUT /funcionarios/{id}`
- `GET /funcionarios/{id}/escalas`
- `GET /funcionarios/{id}/pontos`

### 3.9 Frota

- `GET /veiculos`
- `POST /veiculos`
- `GET /veiculos/{id}`
- `PUT /veiculos/{id}`
- `POST /veiculos/{id}/manutencao`
- `GET /veiculos/{id}/historico`

### 3.10 Inventario

- `GET /inventario/equipamentos`
- `POST /inventario/equipamentos`
- `GET /inventario/kits`
- `POST /inventario/kits`
- `POST /inventario/movimentacoes`
- `GET /inventario/movimentacoes`

### 3.11 Ponto e Localizacao

- `POST /ponto`
- `GET /ponto`
- `POST /localizacoes`
- `GET /localizacoes`

### 3.12 Administracao

- `GET /usuarios`
- `POST /usuarios`
- `GET /perfis`
- `POST /perfis`
- `GET /permissoes`
- `GET /auditoria`
- `GET /parametros`

## 4. Payloads Sugeridos

### 4.1 Criar Cliente

```json
{
  "razao_social": "Empresa Exemplo LTDA",
  "nome_fantasia": "Empresa Exemplo",
  "tipo_pessoa": "juridica",
  "documento": "00.000.000/0001-00",
  "telefone": "(65) 99999-9999",
  "email": "contato@exemplo.com"
}
```

### 4.2 Criar Orcamento

```json
{
  "cliente_id": 1,
  "responsavel_comercial_id": 10,
  "validade": "2026-06-30",
  "tipo_evento": "locacao",
  "forma_pagamento": "boleto",
  "itens": [
    {
      "tipo_item": "kit",
      "item_id": 3,
      "quantidade": 2,
      "valor_unitario": 500.0
    }
  ]
}
```

### 4.3 Criar Ordem de Servico

```json
{
  "orcamento_id": 15,
  "cliente_id": 1,
  "data_agendada": "2026-06-10",
  "hora_evento": "08:00:00",
  "data_instalacao": "2026-06-09",
  "data_prevista_retorno": "2026-06-11",
  "responsavel_operacional_id": 7
}
```

### 4.4 Baixa Financeira

```json
{
  "data_recebimento": "2026-06-12",
  "valor_recebido": 1500.0,
  "meio_pagamento": "pix",
  "referencia": "TX-12345"
}
```

## 5. Regras de API

- toda rota protegida deve exigir autenticacao;
- rotas de escrita devem validar permissao;
- alteracoes criticas devem gerar log;
- status invalidos devem retornar `422`;
- conflitos de agenda ou estoque devem retornar `409`;
- endpoints de listagem devem aceitar pagina, limite e filtro.

## 6. Filtros Padrao

- `status`
- `data_inicio`
- `data_fim`
- `cliente_id`
- `responsavel_id`
- `usuario_id`
- `tipo_evento`
- `numero`

## 7. Ordenacao Padrao

- por data descrescente;
- por numero crescente;
- por status;
- por cliente;
- por vencimento.

## 8. Paginação

Retorno sugerido:

```json
{
  "data": [],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 0
  }
}
```

## 9. Observacao

Esta especificacao foi convertida para OpenAPI / Swagger em [openapi.yaml](/var/home/notNilton/Workspace/nilbyte/erp-leasing/openapi.yaml), com paths, schemas e exemplos iniciais.
