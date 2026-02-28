.PHONY: curl-pod
curl-pod:
	kubectl run mycurlpod --image=curlimages/curl -i --tty -- sh
	kubectl delete pod mycurlpod