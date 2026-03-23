.PHONY: curl-pod
curl-pod:
	kubectl run mycurlpod --image=curlimages/curl -i --tty -- sh
	kubectl delete pod mycurlpod

.PHONY: grafana
grafana:
	kubectl port-forward -n monitoring deployments/prometheus-grafana 3000

.PHONY: prometheus
prometheus:
	kubectl port-forward -n monitoring prometheus-prometheus-kube-prometheus-prometheus-0 9090

.PHONY: connect
connect:
	gcloud container clusters get-credentials gke-standard \
  --zone us-central1-a \
  --project midia-in-da-house