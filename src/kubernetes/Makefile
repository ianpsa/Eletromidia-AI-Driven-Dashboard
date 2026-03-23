include .env
export

.PHONY: curl-pod
curl-pod:
	kubectl run mycurlpod --image=curlimages/curl -i --tty -- sh
	kubectl delete pod mycurlpod

.PHONY: connect
connect:
	gcloud container clusters get-credentials $(CLUSTER_NAME) \
		--zone $(GKE_ZONE) \
		--project $(PROJECT_ID)

.PHONY: auth-plugin
auth-plugin:
	sudo apt-get install google-cloud-cli-gke-gcloud-auth-plugin

.PHONY: node_pools
node_pools:
	gcloud container node-pools list \
	--cluster=$(CLUSTER_NAME) \
	--zone=$(GKE_ZONE) \
	--project=$(PROJECT_ID)

	gcloud container node-pools describe main-pool \
	--cluster=$(CLUSTER_NAME) \
	--zone=$(GKE_ZONE) \
	--project=$(PROJECT_ID)