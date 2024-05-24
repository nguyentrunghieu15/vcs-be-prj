docker build -f ./cmd/auth/Dockerfile -t auth .
docker build -f ./cmd/user/Dockerfile -t user .
docker build -f ./cmd/server/Dockerfile -t server .
docker build -f ./cmd/gateway/Dockerfile -t gateway .
docker build -f ./cmd/mail_sender/Dockerfile -t mail_sender .
docker build -f ./cmd/exporter/Dockerfile -t exporter .
docker build -f ./cmd/file_server/Dockerfile -t file_server .
docker build -f ./cmd/health_check/Dockerfile -t health_check .

