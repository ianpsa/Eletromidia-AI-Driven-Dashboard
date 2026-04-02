# Uso:
#   ./nodepool.sh create   → sobe o nó de teste
#   ./nodepool.sh delete   → remove o nó após o teste
# ─────────────────────────────────────────────────────────────────────────────

# ── Ajuste estas variáveis para o seu ambiente ────────────────────────────────
CLUSTER_NAME="eletromidia-gke-cluster"          # nome do seu cluster GKE
REGION="southamerica-east1-a"                # região ou zona do cluster
PROJECT="midia-in-da-house"          # project ID no GCP
POOL_NAME="loadtest-pool"           # nome do novo node pool
# ─────────────────────────────────────────────────────────────────────────────

case "$1" in

  # ── Criar o node pool de teste ─────────────────────────────────────────────
  create)
    echo "▶ Criando node pool '$POOL_NAME'..."

    gcloud container node-pools create "$POOL_NAME" \
      --cluster="$CLUSTER_NAME" \
      --region="$REGION" \
      --project="$PROJECT" \
      --machine-type="e2-standard-2" \
      --num-nodes=1 \
      --node-taints="dedicated=loadtest:NoSchedule" \
      --node-labels="dedicated=loadtest" \
      --enable-autoupgrade \
      --no-enable-autorepair

    echo "✅ Node pool criado. Aguardando nó ficar Ready..."
    kubectl wait nodes \
      --selector="dedicated=loadtest" \
      --for=condition=Ready \
      --timeout=180s

    echo "✅ Nó pronto! Execute o Job:"
    echo "   kubectl apply -f job.yaml -n kafka"
    echo "   kubectl logs -f job/k6-kafka-loadtest -n kafka"
    ;;

  # ── Deletar o node pool após o teste ──────────────────────────────────────
  delete)
    echo "▶ Removendo node pool '$POOL_NAME'..."

    # Garante que o Job foi removido antes de drenar o nó
    kubectl delete job k6-kafka-loadtest -n kafka --ignore-not-found

    gcloud container node-pools delete "$POOL_NAME" \
      --cluster="$CLUSTER_NAME" \
      --region="$REGION" \
      --project="$PROJECT" \
      --quiet

    echo "✅ Node pool removido. Nenhum custo adicional."
    ;;

  *)
    echo "Uso: $0 {create|delete}"
    exit 1
    ;;
esac