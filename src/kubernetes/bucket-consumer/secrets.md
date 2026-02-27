# Sobre os Segredos do Bucket Consumer

## gcs-sa-key
Para criar o secret, que representa a chave para a Service Account, utilize o código:

```
kubectl create secret generic gcs-sa-key \
> --from-file=cs-api-key.json=/caminho/para/sua-chave.json
```