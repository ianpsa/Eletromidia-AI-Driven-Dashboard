.PHONY: curl-pod
curl-pod:
	kubectl run mycurlpod --image=curlimages/curl -i --tty -- sh
	kubectl delete pod mycurlpod

.PHONY: connect
connect:
	gcloud container clusters get-credentials gke-standard \
  --zone us-central1-a \
  --project midia-in-da-house