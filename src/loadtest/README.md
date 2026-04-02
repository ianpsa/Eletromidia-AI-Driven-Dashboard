# k6 Kafka Load Test — 100k msg/s

Teste de carga no tópico `geodata-dev` usando **k6 + xk6-kafka**.

## Estrutura

```
k6-kafka-loadtest/
├── loadtest.js     # Script k6 com o cenário de carga
├── Dockerfile      # Build multi-stage: compila k6 + xk6-kafka
├── k8s-job.yaml    # Kubernetes Job para rodar no cluster
└── README.md
```

## 1. Build e Push da imagem

```bash
# Substitua pelo seu registry
export REGISTRY=seu-registry.io/seu-projeto

docker build -t $REGISTRY/k6-kafka-loadtest:latest .
docker push $REGISTRY/k6-kafka-loadtest:latest
```

## 2. Atualizar o manifest

Edite `k8s-job.yaml` e substitua o campo `image`:

```yaml
image: seu-registry.io/seu-projeto/k6-kafka-loadtest:latest
```

## 3. Deploy no Kubernetes

```bash
# Aplica o Job no namespace kafka (mesmo do cluster)
kubectl apply -f k8s-job.yaml -n kafka

# Acompanha os logs em tempo real
kubectl logs -f job/k6-kafka-loadtest -n kafka
```

## 4. Verificar resultado

```bash
# Status do Job
kubectl get job k6-kafka-loadtest -n kafka

# Logs completos com métricas
kubectl logs job/k6-kafka-loadtest -n kafka
```

## 5. Limpar após o teste

```bash
kubectl delete job k6-kafka-loadtest -n kafka
```

---

## Parâmetros do cenário

| Parâmetro          | Valor     |
|--------------------|-----------|
| Taxa alvo          | 100.000/s |
| Duração            | 20s       |
| Total de mensagens | ~2.000.000 |
| VUs pré-alocadas   | 200       |
| VUs máximas        | 500       |
| Batch size         | 500       |
| Compressão         | Snappy    |
| Balanceador        | Round Robin (50 partições) |

## Requisitos de recurso (requests/limits no Job)

| Recurso | Request | Limit |
|---------|---------|-------|
| CPU     | 4 cores | 8 cores |
| Memory  | 2Gi     | 4Gi     |