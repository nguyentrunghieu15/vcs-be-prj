package healthcheck

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	beServer "github.com/nguyentrunghieu15/vcs-be-prj/pkg/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
)

type HealthService struct {
	serverRepo     *beServer.ServerRepositoryDecorator
	elasticService *ElasticServiceHeath
}

func NewHealthService() *HealthService {
	dsnPostgres := fmt.Sprintf(
		"host=%v user=%v password=%v dbname=%v port=%v sslmode=%v",
		env.GetEnv("POSTGRES_ADDRESS"),
		env.GetEnv("POSTGRES_USERNAME"),
		env.GetEnv("POSTGRES_PASSWORD"),
		env.GetEnv("POSTGRES_DATABASE"),
		env.GetEnv("POSTGRES_PORT"),
		env.GetEnv("POSTGRES_SSLMODE"),
	)
	postgres, err := managedb.GetConnection(
		managedb.Connection{
			Context: &managedb.PostgreContext{},
			Dsn:     dsnPostgres,
		})
	if err != nil {
		log.Fatalf("Server service : Can't connect to PostgresSQL Database :%v", err)
	}
	log.Println("Connected database")
	connPostgres, _ := postgres.(*gorm.DB)

	elasticConfig := elasticsearch.Config{
		Addresses:              []string{env.GetEnv("ELASTIC_ADDRESS").(string)},
		Username:               env.GetEnv("ELASTICSEARCH_USERNAME").(string),
		Password:               env.GetEnv("ELASTICSEARCH_PASSWORD").(string),
		CertificateFingerprint: env.GetEnv("ELASTIC_CERT_FINGER").(string),
	}

	return &HealthService{serverRepo: beServer.NewServerRepository(connPostgres),
		elasticService: NewElasticServiceHeath(elasticConfig)}
}

func (h *HealthService) Mintor() {
	c := make(chan ServerDocument, 10)
	go h.UpdateStatusServer(c)
	for {
		h.Check(c)
		time.Sleep(10 * time.Minute)
	}
}

func (h *HealthService) Check(c chan ServerDocument) {
	allServer, _ := h.serverRepo.GetAllServers()
	for _, server := range allServer {
		go checkAServer(server, c)
	}
}

func checkAServer(server model.Server, c chan ServerDocument) {
	out, _ := exec.Command("ping", server.Ipv4, "-c 5", "-i 3", "-w 10").Output()
	if strings.Contains(string(out), "Destination Host Unreachable") ||
		strings.Contains(string(out), "100% packet loss") {
		c <- ServerDocument{ID: server.ID,
			Name:   server.Name,
			Ipv4:   server.Ipv4,
			Status: model.Off,
			At:     time.Now(),
			In:     10,
		}
	} else {
		c <- ServerDocument{
			ID:     server.ID,
			Name:   server.Name,
			Ipv4:   server.Ipv4,
			Status: model.On,
			At:     time.Now(),
			In:     10,
		}
	}
}

func (h *HealthService) UpdateStatusServer(c chan ServerDocument) {
	for v := range c {
		err := h.elasticService.CreateDocument(&v)
		h.serverRepo.UpdateOneById(v.ID, map[string]interface{}{
			"status": v.Status,
		})
		if err != nil {
			log.Fatalln(err)
		}
	}
}
