sudo docker build -f ./cmd/auth/Dockerfile -t auth .
sudo docker build -f ./cmd/user/Dockerfile -t user .
sudo docker build -f ./cmd/server/Dockerfile -t server .
sudo docker build -f ./cmd/gateway/Dockerfile -t gateway .
sudo docker build -f ./cmd/mail_sender/Dockerfile -t mail_sender .
sudo docker build -f ./cmd/exporter/Dockerfile -t exporter .
sudo docker build -f ./cmd/file_server/Dockerfile -t file_server .
sudo docker build -f ./cmd/health_check/Dockerfile -t health_check .

