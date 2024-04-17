package main

import (
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/exporter"
)

func main() {
	serverEnv := map[string]env.ConfigEnv{
		"FILE_SERVER_ADDRESS": {IsRequire: true, Type: env.STRING},
		"FILE_SERVER_PORT":    {IsRequire: true, Type: env.STRING},

		"KAFKA_BOOTSTRAP_SERVER": {IsRequire: true, Type: env.STRING},
		"KAFKA_GROUP_ID":         {IsRequire: true, Type: env.STRING},
		"KAFKA_TOPIC_EXPORT":     {IsRequire: true, Type: env.STRING},
	}

	env.Load(".env", serverEnv)

	worker := exporter.NewExporterWorker()
	worker.Work()
}
