.PHONY: gateway auth user server exporter kafka create-topic file_server

gateway:
	go run ./cmd/gateway

auth:
	go run ./cmd/auth
user:
	go run ./cmd/user
server:
	go run ./cmd/server
kafka:
	sudo sh ./script/kafka-run.sh
create-topic:
	sudo sh ./script/kafka-create-topic.sh
redis:
	redis-server
mail_sender:
	go run ./cmd/mail_sender

file_server:
	go run ./cmd/file_server
	
exporter:
	go run ./cmd/exporter
