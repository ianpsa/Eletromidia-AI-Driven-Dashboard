.PHONY: curl-pod
curl-pod:
	kubectl run mycurlpod -n dev --image=curlimages/curl -i --tty -- sh
	kubectl delete pod mycurlpod -n dev