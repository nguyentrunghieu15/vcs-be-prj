package healthcheck

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

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
	return &HealthService{serverRepo: beServer.NewServerRepository(connPostgres)}
}

func (h *HealthService) Mintor() {
	for {
		h.Check()
		time.Sleep(10 * time.Minute)
	}
}

func (h *HealthService) Check() {
	allServer, _ := h.serverRepo.GetAllServers()
	for _, server := range allServer {
		go checkAServer(server)
	}
}

func checkAServer(server model.Server) {
	out, _ := exec.Command("ping", server.Ipv4, "-c 5", "-i 3", "-w 10").Output()
	if strings.Contains(string(out), "Destination Host Unreachable") ||
		strings.Contains(string(out), "100% packet loss") {
		fmt.Println(server.Ipv4, "TANGO DOWN")
	} else {
		fmt.Println(server.Ipv4, "IT'S ALIVEEE")
	}
}
