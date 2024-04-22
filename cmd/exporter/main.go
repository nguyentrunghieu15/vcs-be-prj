package main

import (
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/exporter"
)

func main() {
	exporterEnv := map[string]env.ConfigEnv{
		"FILE_SERVER_ADDRESS": {IsRequire: true, Type: env.STRING},
		"FILE_SERVER_PORT":    {IsRequire: true, Type: env.STRING},

		"KAFKA_BOOTSTRAP_SERVER": {IsRequire: true, Type: env.STRING},
		"KAFKA_GROUP_ID":         {IsRequire: true, Type: env.STRING},
		"KAFKA_TOPIC_EXPORT":     {IsRequire: true, Type: env.STRING},

		"EXPORTER_TEMP_FOLDER": {IsRequire: true, Type: env.STRING},

		"POSTGRES_ADDRESS":  {IsRequire: true, Type: env.STRING},
		"POSTGRES_PORT":     {IsRequire: true, Type: env.INT},
		"POSTGRES_USERNAME": {IsRequire: true, Type: env.STRING},
		"POSTGRES_PASSWORD": {IsRequire: true, Type: env.STRING},
		"POSTGRES_DATABASE": {IsRequire: true, Type: env.STRING},
		"POSTGRES_SSLMODE":  {IsRequire: true, Type: env.STRING},
	}

	env.Load("./cmd/exporter/.env", exporterEnv)

	worker := exporter.NewExporterWorker()
	worker.Work()
}
