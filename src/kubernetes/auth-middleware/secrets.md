# Sobre os Segredos do Auth Middleware

O Auth Middleware precisa de uma Service Account (SA) com as seguintes permissões no GCP:
- `Project IAM Viewer` (`roles/resourcemanager.projectIamViewer`)

Isso permite que o backend consulte as políticas de IAM do projeto para validar as roles dos usuários.

## gcp-sa-key (Opcional)
Se você estiver rodando fora do GCP (ou em um ambiente GKE sem Workload Identity), você precisará fornecer uma chave JSON da SA.

Para criar o secret:

```bash
kubectl create secret generic gcp-sa-key \
  --from-file=gcp-sa-key.json=/caminho/para/sua-chave.json
```

Certifique-se de descomentar as seções de `volumeMounts` e `volumes` no arquivo `deployment.yaml` se for utilizar este método.
