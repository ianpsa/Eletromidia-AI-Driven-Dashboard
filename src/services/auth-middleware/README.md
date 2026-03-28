# Auth Middleware

O Auth Middleware é um serviço em Rust responsável pela autenticação e autorização do sistema, utilizando Firebase Auth para validação de tokens e GCP IAM para gestão de permissões (roles).

## Funcionalidades
- Validação de tokens do Firebase.
- Mapeamento de permissões do GCP IAM para roles da aplicação (`Viewer`, `Editor`, `Admin`).
- Middleware para proteção de rotas por role.
- Endpoints de saúde (`/healthz`) e métricas (`/metrics`).

## Configuração

Utilize o arquivo `.env` para configurar o serviço:

```env
FIREBASE_PROJECT_ID=seu-projeto-firebase
GCP_PROJECT_ID=seu-projeto-gcp

# Mapeamento de Roles (Opcional, possui defaults no código)
ROLE_ADMIN=roles/resourcemanager.projectIamAdmin
ROLE_EDITOR=roles/looker.developer
ROLE_VIEWER=roles/looker.accessUser
```

## Como Rodar Localmente

### Via Rust (Cargo)

```bash
cargo run
```

### Via Docker

```bash
docker build -t auth-middleware .
docker run -p 5000:5000 --env-file .env auth-middleware
```

## Deployment no GCP (GKE)

Os manifestos estão localizados em `kubernetes/auth-middleware/`.

1. Certifique-se de configurar o `ConfigMap` com os IDs corretos.
2. A Service Account do GKE precisa da permissão `Project IAM Viewer`.
3. Aplique os manifestos:

```bash
kubectl apply -f kubernetes/auth-middleware/
```
