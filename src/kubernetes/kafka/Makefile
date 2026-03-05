# Makefile para automações dentro do ecossistema Kafka

.PHONY: producer
producer:
	kubectl -n kafka run kafka-producer \
	-ti --image=quay.io/strimzi/kafka:0.42.0-kafka-3.7.1 \
	--rm=true --restart=Never \
	-- bin/kafka-console-producer.sh --bootstrap-server my-cluster-kafka-bootstrap:9092 --topic geodata

.PHONY: producer-dev
producer-dev:
	kubectl -n kafka run kafka-producer-dev \
	-ti --image=quay.io/strimzi/kafka:0.42.0-kafka-3.7.1 \
	--rm=true --restart=Never \
	-- bin/kafka-console-producer.sh --bootstrap-server my-cluster-kafka-bootstrap:9092 --topic geodata-dev

.PHONY: topic-logs
topic-logs:
	kubectl exec -it my-cluster-dual-role-0 -n kafka -- \
	bin/kafka-console-consumer.sh \
	--bootstrap-server my-cluster-kafka-bootstrap:9092 \
	--topic geodata-dev \
	--from-beginning \
	--property print.partition=true \
	--property print.offset=true \
	--property print.key=true \
	--property print.timestamp=true

.PHONY: get-partitions
get-partitions:
	kubectl exec -it my-cluster-dual-role-0 -n kafka -- \
  bin/kafka-topics.sh \
  --bootstrap-server my-cluster-kafka-bootstrap:9092 \
  --describe \
  --topic geodata