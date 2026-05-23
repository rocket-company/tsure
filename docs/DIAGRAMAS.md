# Diagramas - ERP de Locações e Leasing de Projetos

## 1. Ciclo de Vida da Ordem de Serviço

```mermaid
stateDiagram-v2
    direction TD
    [*] --> Orcamento : criar proposta
    Orcamento --> EmAnalise : enviar para aprovação
    EmAnalise --> Aprovado : gestor aprova
    EmAnalise --> Recusado : gestor reprova
    Recusado --> Orcamento : ajustar e reenviar
    Aprovado --> Agendado : agendar serviço
    Agendado --> Separacao : almoxarifado separa
    Separacao --> Instalado : equipe instala
    Instalado --> EmExecucao : execução em andamento
    EmExecucao --> AguardandoRetorno : evento concluído
    AguardandoRetorno --> Finalizado : retorno conferido
    Finalizado --> [*]

    Orcamento --> Cancelado
    EmAnalise --> Cancelado
    Aprovado --> Cancelado
    Agendado --> Cancelado
    Separacao --> Cancelado
    Instalado --> Cancelado
    EmExecucao --> Cancelado
    AguardandoRetorno --> Cancelado
    Cancelado --> [*]
```

---

## 2. Fluxo Comercial

```mermaid
flowchart TD
    A([Cliente solicita atendimento]) --> B[Comercial cria orçamento]
    B --> C[Adiciona itens, kits e valores]
    C --> D[Define prazo e forma de pagamento]
    D --> E[Envia para aprovação]
    E --> F{Análise do gestor}
    F -->|Aprovado| G[Converte em Ordem de Serviço]
    F -->|Reprovado| H[Devolve com motivo]
    F -->|Prazo expirado| I[Arquiva como expirado]
    H --> J{Comercial revisa?}
    J -->|Sim| C
    J -->|Não| I
    G --> K([Fluxo Operacional])
```

---

## 3. Fluxo Operacional

```mermaid
flowchart TD
    A([OS aprovada]) --> B[Equipe escalada]
    A --> C[Itens reservados no estoque]
    A --> D[Veículo reservado]
    B & C & D --> E[Saída para campo]
    E --> F[Check-in no local]
    F --> G[Execução da instalação]
    G --> H{Instalação OK?}
    H -->|Ocorrência| I[Registra ocorrência]
    I --> G
    H -->|Concluída| J[Registra fotos e observações]
    J --> K[Status: Em execução]
    K --> L[Aguarda evento / prazo]
    L --> M[Retirada dos equipamentos]
    M --> N[Conferência de retorno]
    N --> O{Itens conferidos?}
    O -->|Avaria ou perda| P[Registra ocorrência de avaria]
    P --> Q[Encerra com pendência]
    O -->|Tudo OK| R[Encerramento normal]
    R & Q --> S([Fluxo Financeiro])
```

---

## 4. Fluxo Financeiro

```mermaid
flowchart TD
    A([Serviço encerrado]) --> B[Gera Conta a Receber]
    B --> C[Emite título com vencimento]
    C --> D{Recebimento}
    D -->|Pagamento total| E[Baixa total]
    D -->|Pagamento parcial| F[Baixa parcial]
    D -->|Inadimplência| G[Entra em cobrança]
    F --> H[Saldo em aberto]
    H --> D
    G --> I{Renegociação?}
    I -->|Sim| J[Registra acordo e novas parcelas]
    J --> D
    I -->|Não| K[Mantém inadimplente]
    E --> L([Conciliado])
```

---

## 5. Fluxo Fiscal

```mermaid
flowchart TD
    A([Serviço elegível para fiscal]) --> B[Valida dados obrigatórios]
    B --> C{Dados completos?}
    C -->|Não| D[Solicita correção ao responsável]
    D --> B
    C -->|Sim| E[Emite documento fiscal]
    E --> F[Vincula NF à Ordem de Serviço]
    F --> G[Arquiva XML e PDF]
    G --> H{Necessidade de cancelamento?}
    H -->|Sim| I[Cancela com trilha de auditoria]
    I --> J[Reemissão se necessário]
    J --> E
    H -->|Não| K([Histórico auditável])
```

---

## 6. Fluxo de RH e Campo

```mermaid
flowchart TD
    A([OS agendada]) --> B[RH monta escala de equipe]
    B --> C[Colaborador escalado]
    C --> D[Registro de ponto: entrada]
    D --> E[Localização: check-in no local]
    E --> F[Execução do serviço]
    F --> G[Localização: check-out]
    G --> H[Registro de ponto: saída]
    H --> I[Horas consolidadas]
    I --> J{Horas extras?}
    J -->|Sim| K[Registra HE para folha]
    J -->|Não| L[Produtividade registrada]
    K --> L
    L --> M([Relatório de equipe])
```

---

## 7. Estados do Inventário

```mermaid
stateDiagram-v2
    [*] --> Disponivel : cadastro
    Disponivel --> Reservado : reserva para OS
    Reservado --> Separado : separação almoxarifado
    Separado --> EmCampo : saída para instalação
    EmCampo --> Retornado : retorno confirmado
    Retornado --> Disponivel : conferido e liberado
    EmCampo --> Avariado : dano identificado em campo
    Retornado --> Avariado : avaria detectada no retorno
    Avariado --> Manutencao : enviado para reparo
    Manutencao --> Disponivel : reparo concluído
    Disponivel --> Baixado : descarte
    Avariado --> Baixado : perda total
    Baixado --> [*]
```

---

## 8. Estados da Frota

```mermaid
stateDiagram-v2
    [*] --> Disponivel : cadastro
    Disponivel --> Reservado : reserva para OS
    Reservado --> EmRota : saída com KM inicial
    EmRota --> EmOperacao : chegada no local
    EmOperacao --> EmRota : retorno
    EmRota --> Disponivel : retorno com KM final
    Disponivel --> EmManutencao : manutenção programada
    EmRota --> EmManutencao : pane ou avaria em rota
    EmManutencao --> Disponivel : manutenção concluída
    Disponivel --> Indisponivel : retirado de operação
    Indisponivel --> [*]
```

---

## 9. Mapa de Navegação

```mermaid
flowchart LR
    D([Dashboard])

    D --> COM[Comercial]
    COM --> CLI[Clientes]
    COM --> ORC[Orçamentos]
    COM --> APR[Aprovações]

    D --> OP[Operação]
    OP --> AG[Agenda]
    OP --> OS[Ordens de Serviço]
    OP --> INS[Instalações]
    OP --> RET[Retiradas]
    OP --> OCO[Ocorrências]

    D --> FIN[Financeiro]
    FIN --> CR[Contas a Receber]
    FIN --> REC[Recebimentos]
    FIN --> SAD[Saldos / Inadimplência]

    D --> FSC[Fiscal]
    FSC --> DOC[Documentos Fiscais]
    FSC --> EMI[Emissão]
    FSC --> CAN[Cancelamento / Reemissão]

    D --> REC2[Recursos]
    REC2 --> FUN[Funcionários]
    REC2 --> ESC[Escalas e Ponto]
    REC2 --> LOC[Localização]
    REC2 --> FRO[Frota]
    REC2 --> INV[Inventário]

    D --> ADM[Administração]
    ADM --> USR[Usuários e Perfis]
    ADM --> PER[Permissões]
    ADM --> LOG[Logs de Auditoria]
    ADM --> PAR[Parâmetros]
```

---

## 10. ERD — Domínio Comercial

```mermaid
erDiagram
    CLIENTES ||--o{ CONTATOS : possui
    CLIENTES ||--o{ ENDERECOS : possui
    CLIENTES ||--o{ ORCAMENTOS : solicita

    ORCAMENTOS ||--o{ ORCAMENTO_ITENS : contem
    ORCAMENTOS ||--|| APROVACOES : recebe
    ORCAMENTOS ||--o| ORDENS_SERVICO : converte

    CLIENTES {
        int id PK
        string razao_social
        string nome_fantasia
        string tipo_pessoa
        string documento UK
        string telefone
        string email
        string status
    }

    CONTATOS {
        int id PK
        int cliente_id FK
        string nome
        string cargo
        string telefone
        string email
        bool principal
    }

    ENDERECOS {
        int id PK
        int cliente_id FK
        string tipo
        string logradouro
        string cidade
        string estado
        string cep
        bool principal
    }

    ORCAMENTOS {
        int id PK
        int cliente_id FK
        int responsavel_comercial_id FK
        string numero UK
        date data_orcamento
        date validade
        string status
        string tipo_evento
        string forma_pagamento
        decimal valor_total
        decimal desconto_total
    }

    ORCAMENTO_ITENS {
        int id PK
        int orcamento_id FK
        string tipo_item
        int item_id
        string descricao
        int quantidade
        decimal valor_unitario
        decimal valor_total
    }

    APROVACOES {
        int id PK
        int orcamento_id FK
        int usuario_id FK
        string status
        datetime data_aprovacao
        string numero_aprovacao
        string motivo_reprovacao
    }
```

---

## 11. ERD — Domínio Operacional

```mermaid
erDiagram
    ORDENS_SERVICO ||--o{ AGENDAMENTOS : agenda
    ORDENS_SERVICO ||--o{ INSTALACOES : registra
    ORDENS_SERVICO ||--o{ RETIRADAS : encerra
    ORDENS_SERVICO ||--o{ OCORRENCIAS_OPERACIONAIS : gera
    ORDENS_SERVICO ||--o{ ORDEM_SERVICO_STATUS_HIST : historico
    ORDENS_SERVICO ||--o{ ESCALAS : aloca
    ORDENS_SERVICO ||--o{ MOVIMENTACOES_ESTOQUE : utiliza

    ORDENS_SERVICO {
        int id PK
        int orcamento_id FK
        int cliente_id FK
        string numero UK
        string tipo_evento
        string status
        date data_agendada
        datetime data_instalacao
        date data_prevista_retorno
        int responsavel_operacional_id FK
        string telefone_contato
        int endereco_evento_id FK
        string forma_pagamento
        decimal valor_total
        decimal saldo
    }

    AGENDAMENTOS {
        int id PK
        int ordem_servico_id FK
        datetime data_inicio
        datetime data_fim
        string status
    }

    INSTALACOES {
        int id PK
        int ordem_servico_id FK
        int usuario_responsavel_id FK
        datetime data_inicio
        datetime data_fim
        int localizacao_id FK
    }

    RETIRADAS {
        int id PK
        int ordem_servico_id FK
        int usuario_responsavel_id FK
        string tipo_retirada
        date data_retirada
        date data_retorno
        decimal km_saida
        decimal km_retorno
    }

    OCORRENCIAS_OPERACIONAIS {
        int id PK
        int ordem_servico_id FK
        string tipo_ocorrencia
        string gravidade
        string descricao
        bool resolvida
        datetime data_ocorrencia
    }

    ORDEM_SERVICO_STATUS_HIST {
        int id PK
        int ordem_servico_id FK
        string status_anterior
        string status_novo
        int usuario_id FK
        datetime data_movimento
        string motivo
    }
```

---

## 12. ERD — Domínio Financeiro e Fiscal

```mermaid
erDiagram
    ORDENS_SERVICO ||--o{ CONTAS_RECEBER : gera
    ORDENS_SERVICO ||--o{ DOCUMENTOS_FISCAIS : documenta
    CLIENTES ||--o{ CONTAS_RECEBER : titulares
    CLIENTES ||--o{ DOCUMENTOS_FISCAIS : referencia
    CONTAS_RECEBER ||--o{ RECEBIMENTOS : baixa

    CONTAS_RECEBER {
        int id PK
        int ordem_servico_id FK
        int cliente_id FK
        string numero_titulo UK
        string competencia
        date data_emissao
        date data_vencimento
        decimal valor_original
        decimal valor_baixado
        decimal saldo
        string status
        string forma_recebimento
    }

    RECEBIMENTOS {
        int id PK
        int conta_receber_id FK
        datetime data_recebimento
        decimal valor_recebido
        string meio_pagamento
        string referencia
        int usuario_id FK
    }

    DOCUMENTOS_FISCAIS {
        int id PK
        int ordem_servico_id FK
        int cliente_id FK
        string numero
        string serie
        string tipo_documento
        string status
        date data_emissao
        decimal valor_total
        decimal valor_imposto
        string chave_acesso
        string xml_path
        string pdf_path
    }
```

---

## 13. ERD — Domínio de Recursos (RH, Frota e Inventário)

```mermaid
erDiagram
    FUNCIONARIOS ||--o{ ESCALAS : participa
    FUNCIONARIOS ||--o{ PONTOS : marca
    FUNCIONARIOS ||--o{ LOCALIZACOES : registra

    VEICULOS ||--o{ MANUTENCOES_VEICULOS : recebe

    KITS ||--o{ ITENS_KIT : compoe
    EQUIPAMENTOS ||--o{ ITENS_KIT : integra
    EQUIPAMENTOS ||--o{ MOVIMENTACOES_ESTOQUE : movimenta

    FUNCIONARIOS {
        int id PK
        string nome
        string documento
        string cargo
        string funcao
        string telefone
        string email
        string status
    }

    ESCALAS {
        int id PK
        int funcionario_id FK
        int ordem_servico_id FK
        datetime data_inicio
        datetime data_fim
        string funcao_na_os
        string status
    }

    PONTOS {
        int id PK
        int funcionario_id FK
        int ordem_servico_id FK
        date data_ponto
        time hora_entrada
        time hora_saida
        decimal total_horas
    }

    LOCALIZACOES {
        int id PK
        int funcionario_id FK
        int ordem_servico_id FK
        decimal latitude
        decimal longitude
        datetime data_localizacao
        string origem
    }

    VEICULOS {
        int id PK
        string placa UK
        string modelo
        string marca
        int ano
        decimal km_atual
        string status
        int motorista_padrao_id FK
    }

    MANUTENCOES_VEICULOS {
        int id PK
        int veiculo_id FK
        date data_manutencao
        string tipo
        decimal km
        decimal custo
        string status
    }

    EQUIPAMENTOS {
        int id PK
        string codigo_patrimonio UK
        string nome
        string categoria
        int quantidade_total
        int quantidade_disponivel
        string status
    }

    KITS {
        int id PK
        string nome
        string descricao
        string status
    }

    ITENS_KIT {
        int id PK
        int kit_id FK
        int equipamento_id FK
        int quantidade
    }

    MOVIMENTACOES_ESTOQUE {
        int id PK
        int equipamento_id FK
        int kit_id FK
        int ordem_servico_id FK
        string tipo_movimentacao
        int quantidade
        datetime data_movimentacao
        string origem
        string destino
    }
```

---

## 14. ERD — Domínio de Segurança e Auditoria

```mermaid
erDiagram
    PERFIS ||--o{ PERMISSOES : define
    PERFIS ||--o{ USUARIOS : atribui
    USUARIOS ||--o{ LOGS_AUDITORIA : gera
    USUARIOS ||--o{ ANEXOS : envia

    USUARIOS {
        int id PK
        string nome
        string login UK
        string email UK
        string senha_hash
        string status
        int perfil_id FK
        datetime ultimo_acesso
    }

    PERFIS {
        int id PK
        string nome
        string descricao
        string status
    }

    PERMISSOES {
        int id PK
        int perfil_id FK
        string codigo
        string descricao
        bool ativo
    }

    LOGS_AUDITORIA {
        int id PK
        int usuario_id FK
        string entidade
        int registro_id
        string acao
        string valor_anterior
        string valor_novo
        string motivo
        datetime data_evento
        string ip_origem
    }

    ANEXOS {
        int id PK
        int usuario_id FK
        string entidade
        int registro_id
        string nome_arquivo
        string caminho_arquivo
        string tipo_mime
        datetime data_upload
    }
```

---

## 15. Roadmap por Fases

```mermaid
flowchart LR
    subgraph F1["Fase 1 — Núcleo Operacional"]
        direction TB
        f1a[Clientes]
        f1b[Orçamentos e Aprovação]
        f1c[Ordens de Serviço]
        f1d[Agenda]
        f1e[Inventário básico]
        f1f[Retirada e Retorno]
        f1g[Usuários e Perfis]
    end

    subgraph F2["Fase 2 — Financeiro e Fiscal"]
        direction TB
        f2a[Contas a Receber]
        f2b[Recebimentos e Baixas]
        f2c[Documentos Fiscais]
        f2d[Relatórios Financeiros]
    end

    subgraph F3["Fase 3 — Recursos e Campo"]
        direction TB
        f3a[Funcionários e Escalas]
        f3b[Ponto e Localização]
        f3c[Frota e Manutenção]
        f3d[Auditoria]
    end

    subgraph F4["Fase 4 — Inteligência"]
        direction TB
        f4a[Dashboards]
        f4b[Indicadores]
        f4c[Automações e Alertas]
        f4d[Integrações Externas]
        f4e[Portal do Cliente]
    end

    F1 --> F2 --> F3 --> F4
```
