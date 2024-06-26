package main

import (
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	healthcheck "github.com/nguyentrunghieu15/vcs-be-prj/pkg/health_check"
)

func main() {
	healthEnv := map[string]env.ConfigEnv{
		"POSTGRES_ADDRESS":  {IsRequire: true, Type: env.STRING},
		"POSTGRES_PORT":     {IsRequire: true, Type: env.INT},
		"POSTGRES_USERNAME": {IsRequire: true, Type: env.STRING},
		"POSTGRES_PASSWORD": {IsRequire: true, Type: env.STRING},
		"POSTGRES_DATABASE": {IsRequire: true, Type: env.STRING},
		"POSTGRES_SSLMODE":  {IsRequire: true, Type: env.STRING},

		"ELASTIC_CERT_PATH":      {IsRequire: true, Type: env.STRING},
		"ELASTICSEARCH_USERNAME": {IsRequire: true, Type: env.STRING},
		"ELASTICSEARCH_PASSWORD": {IsRequire: true, Type: env.STRING},
		"ELASTIC_ADDRESS":        {IsRequire: true, Type: env.STRING},
		"ELASTIC_CERT_FINGER":    {IsRequire: true, Type: env.STRING},
	}

	env.Load("./cmd/health_check/.env", healthEnv)

	doctor := healthcheck.NewHealthService()
	doctor.Mintor()
	c := make(chan byte)
	<-c
}
