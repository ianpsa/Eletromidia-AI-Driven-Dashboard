# Chave de Service Account GCP

> Somente para desenvolvimento local (minikube).

Em produção (GKE), o pod autentica via Workload Identity — nenhum arquivo de chave é necessário.

## Por que é necessária localmente

O minikube não tem identidade GCP. Sem a chave, o cliente BigQuery não consegue autenticar e o serviço falha ao iniciar. A chave é de uma Service Account com permissão `bigquery.dataEditor` no dataset `eletromidia_bq`.

## Como obter

Peça ao responsável pelo projeto GCP (`midia-in-da-house`) para gerar via Console ou CLI:

```bash
gcloud iam service-accounts keys create sa-key.json \
  --iam-account=bq-data-editor@midia-in-da-house.iam.gserviceaccount.com \
  --project=midia-in-da-house
```

## Criar o secret no Kubernetes

Com o arquivo `sa-key.json` em mãos (na raiz do repositório):

```bash
kubectl create secret generic gcp-sa-key \
  --from-file=key.json=./sa-key.json
```

## Verificar

```bash
kubectl get secret gcp-sa-key -o yaml
```

## Importante

- `sa-key.json` está no `.gitignore` — nunca commitar esse arquivo.
- Para revogar uma chave comprometida: Console GCP → IAM → Service Accounts → `bq-data-editor` → Keys → Delete.
