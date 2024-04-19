package healthcheck

import (
	"context"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
)

type ServerDocument struct {
	ID     uuid.UUID
	Name   string
	Ipv4   string
	Status model.ServerStatus
	At     time.Time
	In     time.Duration
}

type ElasticServiceHeath struct {
	client *elasticsearch.TypedClient
	index  string
}

func NewElasticServiceHeath(config elasticsearch.Config) *ElasticServiceHeath {
	typedClient, err := elasticsearch.NewTypedClient(config)
	if err != nil {
		log.Fatalln("Can't create elastic client", err)
	}
	return &ElasticServiceHeath{client: typedClient}
}

func (e *ElasticServiceHeath) CreateDocument(s *ServerDocument) error {
	_, err := e.client.Index(e.index).
		Id(s.ID.String() + s.At.Format(time.RFC3339)).
		Document(s).Do(context.TODO())
	return err
}
