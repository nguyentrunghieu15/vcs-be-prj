package server

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"github.com/segmentio/kafka-go"
)

func ConvertStatusServerModelToStatusServerProto(server model.ServerStatus) pb.ServerStatus {
	switch server {
	case model.On:
		return pb.ServerStatus_ON
	case model.Off:
		return pb.ServerStatus_OFF
	default:
		return pb.ServerStatus_STATUSNONE
	}
}

func ConvertServerModelToServerProto(server model.Server) *pb.Server {
	return &pb.Server{
		Id:        server.ID.String(),
		CreatedAt: server.CreatedAt.String(),
		CreatedBy: int64(server.CreatedBy),
		UpdatedAt: server.UpdatedAt.String(),
		UpdatedBy: int64(server.UpdatedBy),
		Name:      server.Name,
		Status:    ConvertStatusServerModelToStatusServerProto(server.Status),
		Ipv4:      server.Ipv4,
	}
}

func ConvertListServerModelToListServerProto(s []model.Server) []*server.Server {
	var result []*server.Server = make([]*server.Server, 0)
	for _, v := range s {
		result = append(result, ConvertServerModelToServerProto(v))
	}
	return result
}

func ValidateListServerQuery(req *server.ListServerRequest) error {
	if req.GetPagination() != nil {
		if limit := req.GetPagination().Limit; limit != nil && *limit < 1 {
			return fmt.Errorf("Limit must be a positive number")
		}

		if page := req.GetPagination().Page; page != nil && *page < 1 {
			return fmt.Errorf("Page must be a positive number")
		}

		if pageSize := req.GetPagination().PageSize; pageSize != nil && *pageSize < 1 {
			return fmt.Errorf("Page size must be a positive number")
		}

		if sort := req.GetPagination().Sort; sort != nil &&
			*sort != server.TypeSort_ASC &&
			*sort != server.TypeSort_DESC &&
			*sort != server.TypeSort_NONE {
			return fmt.Errorf("Invalid type order")
		}
	}
	return nil
}

func ValidateServerFormMap(server map[string]interface{}) error {
	keyValid := []string{"id", "name", "ipv4", "status"}
	for k, v := range server {
		switch k {
		case keyValid[0]:
			_, e := uuid.Parse(v.(string))
			if e != nil {
				return fmt.Errorf("Id:%v  not is uuid format ", v)
			}
		case keyValid[1]:
		case keyValid[2]:
			ip := net.ParseIP(v.(string))
			if ip == nil {
				return fmt.Errorf("IP address is not valid")
			}

			if ip.To4() == nil {
				return fmt.Errorf("IP address is not v4")
			}
		case keyValid[3]:
			if v.(string) != "on" && v.(string) != "off" {
				return fmt.Errorf("Status is valid")
			}
		default:
			return fmt.Errorf("Only accept key in %v", keyValid)
		}
	}
	return nil
}

func ParseExportRequestToKafkaMessage(req *pb.ExportServerRequest) (*kafka.Message, error) {
	jsonRequest, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("Can'nt parse export to kafka message,%v", err)
	}
	return &kafka.Message{
		Key:   []byte(req.File.GetFileName()),
		Value: jsonRequest,
	}, nil
}
