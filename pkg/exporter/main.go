package exporter

import (
	"context"
	"fmt"
	"log"

	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/server"
)

type ExporterWorker struct {
	kafkaClientConfig server.ConfigConsumer
	kafkaClient       server.ConsumerClientInterface
}

func NewExporterWorker() *ExporterWorker {
	kafkaClientConfig := server.ConfigConsumer{
		Broker:     []string{env.GetEnv("KAFKA_BOOTSTRAP_SERVER").(string)},
		Topic:      env.GetEnv("KAFKA_TOPIC_EXPORT").(string),
		GroupID:    env.GetEnv("KAFKA_GROUP_ID").(string),
		Logger:     log.New(log.Writer(), "Worker", log.Flags()),
		ErrorLoger: log.New(log.Writer(), "Worker", log.Flags()),
	}
	return &ExporterWorker{kafkaClientConfig: kafkaClientConfig}
}

func (e *ExporterWorker) Work() {
	e.kafkaClient = server.NewComsumer(e.kafkaClientConfig)
	for {
		m, err := e.kafkaClient.ReadMessage(context.Background())
		if err != nil {
			fmt.Println(e.kafkaClientConfig)
			fmt.Println("Error", err)
			break
		}
		fmt.Printf(
			"message at topic/partition/offset %v/%v/%v: %s = %s\n",
			m.Topic,
			m.Partition,
			m.Offset,
			string(m.Key),
			string(m.Value),
		)
	}

	if err := e.kafkaClient.CloseReader(); err != nil {
		log.Fatal("failed to stop worker:", err)
	}
}

func (e *ExporterWorker) ExportToFileLocal() {
}

func (e *ExporterWorker) SendToFileServer() {
}
