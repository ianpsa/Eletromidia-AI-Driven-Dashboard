.PHONY: pods
pods:
	kubectl auth can-i list pods --namespace=dev --as=ian.simao@sou.inteli.edu.br

.PHONY: nodes
nodes:
	kubectl auth can-i list nodes --namespace=dev --as=ian.simao@sou.inteli.edu.br

.PHONY: configmaps
configmaps:
	kubectl auth can-i list configmaps --namespace=dev --as=ian.simao@sou.inteli.edu.br

.PHONY: secrets
secrets:
	kubectl auth can-i list secrets --namespace=dev --as=ian.simao@sou.inteli.edu.br

.PHONY: prod
prod:
	kubectl auth can-i list pods --namespace=default --as=ian.simao@sou.inteli.edu.br

.PHONY: install-prometheus
install-prometheus:
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo add stable https://kubernetes-charts.storage.googleapis.com/
	helm repo update
	helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring