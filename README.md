# VCServer

## Description
The repository is source for backend project VCS checkpoint. The server status management system (VCS-SMS) of Company VCS is a robust solution designed to manage the On/Off status of a large fleet of servers, totaling approximately 10,000 units. This repo is microservice structure

## Installing
### Pre- requirement
Before install project. You need to install some requirement below:
```
Required:
    go >= 1.21
    postgreSQL >= 16
    kafka >= 3.17
    redis >= 7.0
    elasticsearch >= 2.7
    nodejs >= 18
```

### Install
1. Clone repo
```
git clone https://github.com/nguyentrunghieu15/vcs-be-prj.git
```

2. Install dependency go package
```
go mod tidy
```
3. Config environment
```
You need to create a file named .env in each folder below:
cmd
├───auth
├───exporter
├───file_server
├───gateway
├───health_check
├───mail_sender
├───server
└───user

For each .env, you need config .env accordingly main.go in same folder
Example:

This is environment variable for gateway:

func main() {
	gatewayConfigEnv := map[string]env.ConfigEnv{
		"JWT_SECRETKEY": {IsRequire: true, Type: env.STRING},

		"ADMIN_EMAIL":    {IsRequire: true, Type: env.STRING},
		"ADMIN_PASSWORD": {IsRequire: true, Type: env.STRING},
		"ADMIN_FULLNAME": {IsRequire: true, Type: env.STRING},

		"GATEWAY_PORT":          {IsRequire: true, Type: env.INT},
		"GATEWAY_ADDRESS":       {IsRequire: true, Type: env.STRING},
		"GATEWAY_LOG_PATH":      {IsRequire: true, Type: env.STRING},
		"GATEWAY_NAME_FILE_LOG": {IsRequire: true, Type: env.STRING},

		"AUTH_PORT":    {IsRequire: true, Type: env.INT},
		"AUTH_ADDRESS": {IsRequire: true, Type: env.STRING},

		"USER_PORT":    {IsRequire: true, Type: env.INT},
		"USER_ADDRESS": {IsRequire: true, Type: env.STRING},

		"SERVER_PORT":    {IsRequire: true, Type: env.INT},
		"SERVER_ADDRESS": {IsRequire: true, Type: env.STRING},

		"POSTGRES_ADDRESS":  {IsRequire: true, Type: env.STRING},
		"POSTGRES_PORT":     {IsRequire: true, Type: env.INT},
		"POSTGRES_USERNAME": {IsRequire: true, Type: env.STRING},
		"POSTGRES_PASSWORD": {IsRequire: true, Type: env.STRING},
		"POSTGRES_DATABASE": {IsRequire: true, Type: env.STRING},
		"POSTGRES_SSLMODE":  {IsRequire: true, Type: env.STRING},

		"GATEWAY_UPLOAD_FOLDER": {IsRequire: true, Type: env.STRING},

		"FILE_SERVER_ADDRESS": {IsRequire: true, Type: env.STRING},
		"FILE_SERVER_PORT":    {IsRequire: true, Type: env.STRING},

		"MAIL_SENDER_ADDRESS": {IsRequire: true, Type: env.STRING},
		"MAIL_SENDER_PORT":    {IsRequire: true, Type: env.INT},
	}

```

## Running
### Before running
<p>Running postgreSQL ,redis server, elsticsearch, kafka</p> 
<p>Create kafka topic export_file . You can chose num partition and replica for yourself.</p>

## Running 
1. Run gateway
```
go run ./cmd/gateway
```
2. Run auth service
```
go run ./cmd/auth
```
3. Run user service
```
go run ./cmd/user
```
4. Run server service
```
go run ./cmd/server
```
5. Run health check service
```
go run ./cmd/health_check
```
6. Run file server
```
go run ./cmd/file_server
```
7. Run mail sender service
```
go run ./cmd/mail_sender
```

8. Run run exporter
```
You can run more 1 exporter , exporter is a simple comsumer kafka which read topic export_file 

go run ./cmd/exporter
```
Notice: You can run each service on another machine physic

## Open API
You can read API endtrie point of gateway in static folder
- auth.swagger.json
- mail_sender.swagger.json
- server.swagger.json
- user.swagger.json