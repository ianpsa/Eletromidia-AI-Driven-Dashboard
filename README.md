# Midia in da house

## Grupo - Midia in da house

<p align="center">
  <img src="assets/G05%20-%20Sprint%205.png" alt="Logo Midia in da house" width="600">
</p>

## Integrantes:

<div align="center">
  <table>
    <tr>
      <td align="center">
        <a href="https://www.linkedin.com/in/davi-abreu-da-silveira/">
          <img src="assets/davi.jpg" style="border-radius: 10%; width: 150px;" alt="Davi Abreu da Silveira"/><br>
          <sub><b>Davi Abreu da Silveira</b></sub>
        </a>
      </td>
      <td align="center">
        <a href="https://www.linkedin.com/in/ian-pereira-simao/">
          <img src="assets/ian.jpg" style="border-radius: 10%; width: 150px;" alt="Ian Pereira Simão"/><br>
          <sub><b>Ian Pereira Simão</b></sub>
        </a>
      </td>
      <td align="center">
        <a href="https://www.linkedin.com/in/julia-lika-ishikawa/">
          <img src="assets/lika.png" style="border-radius: 10%; width: 150px;" alt="Júlia Lika Ishikawa"/><br>
          <sub><b>Júlia Lika Ishikawa</b></sub>
        </a>
      </td>
      <td align="center">
        <a href="https://www.linkedin.com/in/lucas-periquito-costa/">
          <img src="assets/cuca.jpg" style="border-radius: 10%; width: 150px;" alt="Lucas Periquito Costa"/><br>
          <sub><b>Lucas Periquito Costa</b></sub>
        </a>
      </td>
      <td align="center">
        <a href="https://www.linkedin.com/in/murilo-couto-oliveira/">
          <img src="assets/murilo.jpg" style="border-radius: 10%; width: 150px;" alt="Murilo Couto Oliveira"/><br>
          <sub><b>Murilo Couto Oliveira</b></sub>
        </a>
      </td>
      <td align="center">
        <a href="https://www.linkedin.com/in/yasmim-passos">
          <img src="assets/yas.png" style="border-radius: 10%; width: 150px;" alt="Yasmim Marly Passos"/><br>
          <sub><b>Yasmim Marly Passos</b></sub>
        </a>
      </td>
    </tr>
  </table>
</div>

## Professores:

### Orientador(a)

- [Rodrigo Nicola](https://www.linkedin.com/in/rodrigo-mangoni-nicola-537027158/)

### Instrutores

- [André Godoi Chiovato](https://www.linkedin.com/in/andregodoichiovato/)
- [Francisco de Souza Escobar](https://www.linkedin.com/in/francisco-escobar/)
- [Marcelo de Paula do Desterro Gorini](https://www.linkedin.com/in/marcelodesterro/)
- [Marcelo Luiz do Amaral Gonçalves](https://www.linkedin.com/in/marcelo-gon%C3%A7alves-phd/)
- [Murilo Zanini de Carvalho](https://www.linkedin.com/in/murilo-zanini-de-carvalho-0980415b/)

## Descrição

**Midia in da house** é uma plataforma de planejamento de campanhas de mídia OOH (_Out-of-Home_) assistida por IA, desenvolvida em parceria com a **Eletromidia**. A solução permite que planejadores de mídia identifiquem os melhores pontos de exibição a partir de dados reais de fluxo de pessoas e inventário de telas.

A arquitetura é orientada a dados e composta por:

- **Pipeline de ingestão**: CSVs são processados por um produtor em Rust, publicados em **Kafka** e consumidos para carga no **BigQuery**.
- **Visualização geoespacial**: frontend em React + Deck.GL + MapLibre que renderiza camadas de _hexbins_ sobre o mapa para análise de densidade e cobertura.
- **Agente conversacional**: um _AI agent_ baseado em LangGraph que responde perguntas sobre campanhas, cruzando dados do BigQuery com contexto do usuário.
- **Dashboard analítico**: integração com Looker Studio para relatórios consolidados.

## Estrutura de Pastas

```
g05/
├── data/                   # Datasets CSV (claro, eletromidia, fluxo_claro)
├── docs/                   # Documentação (Next.js + Fumadocs)
├── releases/               # Notas de cada sprint
└── src/
    ├── docker-compose.yml  # Orquestração local (Kafka, agent, auth, frontend)
    ├── frontend/           # React + Vite + Deck.GL + MapLibre
    ├── iac/                # Terraform modules (GCP)
    ├── kubernetes/         # Manifestos Kubernetes por serviço
    └── services/
        ├── agent/              # AI agent (Python, FastAPI, LangGraph)
        ├── auth-middleware/    # Autenticação (Rust, Axum)
        ├── bff-storage/        # BFF para GCS/BigQuery (Go)
        ├── bigquery-writer/    # Consumer Kafka → BigQuery (Go)
        ├── bucket-consumer/    # Eventos GCS → Kafka (Go)
        ├── populate-bigquery/  # Carga inicial do BigQuery (Go)
        └── rust-producer/      # Ingestão de CSV para Kafka (Rust, Actix)
```

## Execução do Projeto

### Documentação

Para executar a documentação localmente:

```bash
git clone https://git.inteli.edu.br/graduacao/2026-1a/t12/g05
cd docs
npm install
npm run dev
```

### Execução Completa

A stack completa (Kafka, frontend, agente de IA e serviços de suporte) é orquestrada via Docker Compose. Pré-requisitos: `git`, `docker` e `docker compose`.

Antes de subir o ambiente, você vai precisar de:

- **API key de um provedor de LLM** (ex.: `GROQ_API_KEY`) para o agente e o frontend.
- **Service Account do BigQuery** (JSON) usada pelo agente e pelo `populate-bigquery` para consultar/popular as tabelas.
- **Service Account do Google Cloud Storage** (JSON, ex.: `cs-api-key.json`) usada pelo `bucket-consumer` e pelo `populate-bigquery`.
- **Projeto Firebase** (ID do projeto) para o `auth-middleware`.
- Arquivos `.env` preenchidos em cada serviço — use os `.env.example` como base (`services/agent`, `services/auth-middleware`, `services/bucket-consumer`, `services/populate-bigquery`, `services/rust-producer/services/csv-kafka-producer`, `frontend`).

Consulte a [documentação oficial do projeto](#documentação) para mais informações sobre como obter cada credencial e quais variáveis são obrigatórias por serviço.

Com as credenciais no lugar, suba a stack:

```bash
git clone https://git.inteli.edu.br/graduacao/2026-1a/t12/g05
cd g05/src
docker compose up --build
```

Após a inicialização, o frontend fica disponível em `http://localhost:5173` e o agente em `http://localhost:8001`.

## Histórico de Lançamentos

### 0.5.0 - Sprint 5

### 0.4.0 - Sprint 4

- Pipeline CI/CD e observabilidade com Prometheus em todos os serviços.
- Visualização por _hexbins_ com sidebar de chat integrada ao frontend.
- Integração BigQuery ↔ Looker Studio, fluxo de autenticação completo e cobertura de testes.

### 0.3.0 - Sprint 3

- Frontend React com telas de Login, Dashboard, Cloud e Agent.
- AI agent baseado em LangGraph para análise conversacional das campanhas.
- _Hardening_ dos manifestos Kubernetes, CSV producer com _health_ e _metrics_ e provisionamento GCP via Terraform.

### 0.2.0 - Sprint 2

- Pipeline de ingestão CSV → Kafka → BigQuery em funcionamento ponta a ponta.
- Primeiros manifestos Kubernetes dos serviços de dados.
- Dashboard inicial em Looker Studio sobre a base consolidada.

### 0.1.0 - Sprint 1

- Fundação do projeto e definição da arquitetura de microserviços.
- Documentação inicial publicada com Next.js + Fumadocs.

## Licença

<img style="height:22px!important;margin-left:3px;vertical-align:text-bottom;" src="https://mirrors.creativecommons.org/presskit/icons/cc.svg?ref=chooser-v1"><img style="height:22px!important;margin-left:3px;vertical-align:text-bottom;" src="https://mirrors.creativecommons.org/presskit/icons/by.svg?ref=chooser-v1"><p xmlns:cc="http://creativecommons.org/ns#" xmlns:dct="http://purl.org/dc/terms/"><a property="dct:title" rel="cc:attributionURL" href="https://git.inteli.edu.br/graduacao/2026-1a/t12/g05">Midia in da house</a> by <a rel="cc:attributionURL dct:creator" property="cc:attributionName" href="https://www.inteli.edu.br/">Inteli</a>, Davi Abreu da Silveira, Ian Pereira Simão, Júlia Lika Ishikawa, Lucas Periquito Costa, Murilo Couto Oliveira, Yasmim Marly Passos is licensed under <a href="http://creativecommons.org/licenses/by/4.0/?ref=chooser-v1" target="_blank" rel="license noopener noreferrer" style="display:inline-block;">Attribution 4.0 International</a>.</p>
