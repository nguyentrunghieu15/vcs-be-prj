package server

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"github.com/segmentio/kafka-go"
)

func ConvertStatusServerModelToStatusServerProto(server model.ServerStatus) pb.Server_ServerStatus {
	switch server {
	case model.On:
		return pb.Server_ON
	case model.Off:
		return pb.Server_OFF
	default:
		return pb.Server_NONE
	}
}

func ConvertServerModelToServerProto(server model.Server) (*pb.Server, error) {
	t, err := json.Marshal(server)
	if err != nil {
		return nil, err
	}

	mapTemp := make(map[string]interface{})

	if err = json.Unmarshal(t, &mapTemp); err != nil {
		return nil, err
	}

	switch server.Status {
	case model.On:
		mapTemp["status"] = pb.Server_ON
	case model.Off:
		mapTemp["status"] = pb.Server_OFF
	default:
		mapTemp["status"] = pb.Server_NONE
	}

	t, err = json.Marshal(mapTemp)
	if err != nil {
		return nil, err
	}

	var result pb.Server
	err = json.Unmarshal(t, &result)
	if err != nil {
		return nil, err
	}

	if _, ok := mapTemp["createdAt"]; ok && mapTemp["createdAt"] != "0001-01-01T00:00:00Z" {
		result.CreatedAt = server.CreatedAt.Format(time.RFC3339)
	}
	if _, ok := mapTemp["updatedAt"]; ok && mapTemp["updatedAt"] != "0001-01-01T00:00:00Z" {
		result.UpdatedAt = server.UpdatedAt.Format(time.RFC3339)
	}
	if _, ok := mapTemp["deletedAt"]; ok && mapTemp["deletedAt"] != nil {
		result.DeletedAt = server.DeletedAt.Time.Format(time.RFC3339)
	}

	if _, ok := mapTemp["createdBy"]; ok {
		result.CreatedBy = int64(server.CreatedBy)
	}
	if _, ok := mapTemp["updatedBy"]; ok {
		result.UpdatedBy = int64(server.UpdatedBy)
	}
	if _, ok := mapTemp["deletedBy"]; ok {
		result.DeletedBy = int64(server.DeletedBy)
	}

	return &result, nil

}

func ConvertListServerModelToListServerProto(s []model.Server) ([]*server.Server, error) {
	var result []*server.Server = make([]*server.Server, 0)
	for _, v := range s {
		t, err := ConvertServerModelToServerProto(v)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
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

func ParseMapCreateServerRequest(req *pb.CreateServerRequest) (map[string]interface{}, error) {
	t, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	mapRequest := make(map[string]interface{})
	result := make(map[string]interface{})

	if err = json.Unmarshal(t, &mapRequest); err != nil {
		return nil, err
	}

	for i := 0; i < len(DefinedFieldCreateServerRequest); i++ {
		if value, ok := mapRequest[DefinedFieldCreateServerRequest[i]["fieldNameProto"]]; ok {
			result[DefinedFieldCreateServerRequest[i]["fieldNameModel"]] = value
		}
	}

	if _, ok := result["Status"]; ok {
		if req.GetStatus() == pb.CreateServerRequest_OFF {
			result["Status"] = model.Off
		}
		if req.GetStatus() == pb.CreateServerRequest_ON {
			result["Status"] = model.On
		}
	}
	return result, nil
}

func ParseMapUpdateServerRequest(req *pb.UpdateServerRequest) (map[string]interface{}, error) {
	t, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	mapRequest := make(map[string]interface{})
	result := make(map[string]interface{})

	if err = json.Unmarshal(t, &mapRequest); err != nil {
		return nil, err
	}

	for i := 0; i < len(DefinedFieldUpdateServerRequest); i++ {
		if value, ok := mapRequest[DefinedFieldUpdateServerRequest[i]["fieldNameProto"]]; ok {
			result[DefinedFieldUpdateServerRequest[i]["fieldNameModel"]] = value
		}
	}

	// if _, ok := result["Status"]; ok {
	// 	if req.GetStatus() == pb.ServerStatus_OFF {
	// 		result["Status"] = model.Off
	// 	}
	// 	if req.GetStatus() == pb.ServerStatus_ON {
	// 		result["Status"] = model.On
	// 	}
	// }
	result["Status"] = nil
	return result, nil
}

func ConvertServerModelMapToServerProto(server map[string]interface{}) (*pb.Server, error) {
	// Convert the map to JSON
	jsonData, _ := json.Marshal(server)
	var structData model.Server
	json.Unmarshal(jsonData, &structData)
	return ConvertServerModelToServerProto(structData)
}

func ConvertListServerModelMapToListServerProto(s []map[string]interface{}) ([]*pb.Server, error) {
	var result []*pb.Server = make([]*pb.Server, 0)
	for _, v := range s {
		t, err := ConvertServerModelMapToServerProto(v)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}
