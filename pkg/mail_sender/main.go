package mailsenderservice

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/go-co-op/gocron/v2"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/auth"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/env"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/logger"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/server"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/mail_sender"
	pbServer "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/managedb"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type LogMessageMailSender map[string]interface{}

type MailSenderServer struct {
	pb.MailServerServer
	elasticService *ElasticService
	l              *logger.LoggerDecorator
	serverRepo     *server.ServerRepositoryDecorator
	authorize      *auth.Authorizer
}

func NewMailSenderServer() *MailSenderServer {
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
	connPostgres.Config.Logger = gormLogger.Default.LogMode(gormLogger.Info)
	newLogger := logger.NewLogger()
	newLogger.Config = logger.LoggerConfig{
		IsLogRotate:     true,
		PathToLog:       env.GetEnv("MAIL_SENDER_LOG_PATH").(string),
		FileNameLogBase: env.GetEnv("MAIL_SENDER_NAME_FILE_LOG").(string),
	}
	elasticConfig := elasticsearch.Config{
		Addresses:              []string{env.GetEnv("ELASTIC_ADDRESS").(string)},
		Username:               env.GetEnv("ELASTICSEARCH_USERNAME").(string),
		Password:               env.GetEnv("ELASTICSEARCH_PASSWORD").(string),
		CertificateFingerprint: env.GetEnv("ELASTIC_CERT_FINGER").(string),
	}
	return &MailSenderServer{
		serverRepo:     server.NewServerRepository(connPostgres),
		elasticService: NewElasticService(elasticConfig),
		l:              newLogger,
		authorize:      &auth.Authorizer{},
	}
}

type ResultStatisticServer struct {
	To              string
	TotalServer     int64
	NumServerOff    int64
	NumServerOn     int64
	AvgServerTimeup float64
}

func (m *MailSenderServer) SendStatisticServerToEmail(ctx context.Context, req *pb.RequestSendStatisticServerToEmail) (*emptypb.Empty, error) {
	m.l.Log(
		logger.INFO,
		LogMessageMailSender{
			"Action": "Invoked  Send Statistic Server To Email",
			"Email":  req.Email,
		},
	)

	fmt.Println(req)

	// Authorize

	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		m.l.Log(
			logger.ERROR,
			LogMessageMailSender{
				"Action": "Send Statistic Server To Email",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	role, ok := header["role"]
	if !ok {
		m.l.Log(
			logger.ERROR,
			LogMessageMailSender{
				"Action": "Send Statistic Server To Email",
				"Error":  "Can't get header from request",
			},
		)
		return nil, status.Error(codes.Internal, "Can't get header from request")
	}

	if !m.authorize.HavePermisionToSendMail(model.UserRole(role[0])) {
		m.l.Log(
			logger.ERROR,
			LogMessageMailSender{
				"Action": "Send Statistic Server To Email",
				"Error":  "Permission denie",
			},
		)
		return nil, status.Error(codes.PermissionDenied, "Can't Send Statistic Server To Email")
	}

	// validate data
	if err := req.Validate(); err != nil {
		m.l.Log(
			logger.ERROR,
			LogMessageMailSender{
				"Action": "Send Statistic Server To Emair",
				"Email":  req.Email,
				"Error":  "Bad request",
				"Detail": err,
			},
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var result ResultStatisticServer

	result.TotalServer, _ = m.serverRepo.CountServers(nil, nil)

	result.NumServerOn, _ = m.serverRepo.CountServers(nil,
		&pbServer.FilterServer{Status: pbServer.ServerStatus_ON.Enum()})

	result.NumServerOff, _ = m.serverRepo.CountServers(nil,
		&pbServer.FilterServer{Status: pbServer.ServerStatus_OFF.Enum()})

	avgResponse, _ := m.elasticService.GetAverageUpTimeServer(req)

	result.AvgServerTimeup = float64(avgResponse.Aggregations["avg_time_up"].(*types.SimpleValueAggregate).Value)
	// Sender data.
	from := env.GetEnv("MAIL_SENDER_EMAIL").(string)
	password := env.GetEnv("MAIL_SENDER_PASSWORD").(string)

	// Receiver email address.
	to := []string{
		req.To,
	}

	// smtp server configuration.
	smtpHost := env.GetEnv("MAIL_SENDER_SMTP_HOST").(string)
	smtpPort := env.GetEnv("MAIL_SENDER_SMTP_PORT").(string)

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	t, _ := template.ParseFiles(env.GetEnv("MAIL_SENDER_TEMPLATE").(string))

	var body bytes.Buffer

	body.Write([]byte(fmt.Sprintf("Subject: Report VCServer")))

	t.Execute(&body, result)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return nil, nil
}

func (m *MailSenderServer) DailySendMail(req *pb.RequestSendStatisticServerToEmail) error {
	m.l.Log(
		logger.INFO,
		LogMessageMailSender{
			"Action": "Daily send email",
			"Email":  req.Email,
		},
	)

	var result ResultStatisticServer

	result.TotalServer, _ = m.serverRepo.CountServers(nil, nil)

	result.NumServerOn, _ = m.serverRepo.CountServers(nil,
		&pbServer.FilterServer{Status: pbServer.ServerStatus_ON.Enum()})

	result.NumServerOff, _ = m.serverRepo.CountServers(nil,
		&pbServer.FilterServer{Status: pbServer.ServerStatus_OFF.Enum()})

	avgResponse, _ := m.elasticService.GetAverageUpTimeServer(req)

	result.AvgServerTimeup = float64(avgResponse.Aggregations["avg_time_up"].(*types.SimpleValueAggregate).Value)
	// Sender data.
	from := env.GetEnv("MAIL_SENDER_EMAIL").(string)
	password := env.GetEnv("MAIL_SENDER_PASSWORD").(string)

	// Receiver email address.
	to := []string{
		req.To,
	}

	// smtp server configuration.
	smtpHost := env.GetEnv("MAIL_SENDER_SMTP_HOST").(string)
	smtpPort := env.GetEnv("MAIL_SENDER_SMTP_PORT").(string)

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	t, _ := template.ParseFiles(env.GetEnv("MAIL_SENDER_TEMPLATE").(string))

	var body bytes.Buffer

	body.Write([]byte(fmt.Sprintf("Subject: Report VCServer")))

	t.Execute(&body, result)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
	if err != nil {
		return fmt.Errorf("Error when send", err)
	}
	return nil
}

func (m *MailSenderServer) WorkDaily() {
	s, _ := gocron.NewScheduler()
	defer func() { _ = s.Shutdown() }()

	_, _ = s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(14, 0, 0),
			),
		),
		gocron.NewTask(
			func(emailSupperAdmin string) {
				err := m.DailySendMail(&pb.RequestSendStatisticServerToEmail{
					From:  time.Now().AddDate(0, 0, -1).Format(time.DateOnly),
					To:    time.Now().AddDate(0, 0, -1).Format(time.DateOnly),
					Email: emailSupperAdmin,
				})
				if err != nil {
					log.Println("Daily Email:", err)
				}
			},
			env.GetEnv("MAIL_SENDER_EMAIL_SUPPER_ADMIN").(string),
		),
	)
	s.Start()
	c := make(chan byte)
	<-c
}
