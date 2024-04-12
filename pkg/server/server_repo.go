package server

import (
	"encoding/json"
	"strings"

	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
)

type ServerRepositoryDecorator struct {
	*model.ServerRepository
	db *gorm.DB
}

func TypeSortToString(v model.TypeSort) string {
	switch v {
	case model.ASC, model.NONE:
		return ""
	case model.DESC:
		return "desc"
	}
	return ""
}

func (u *ServerRepositoryDecorator) FindServers(filter model.FilterQueryInterface) ([]model.Server, error) {
	var server []model.Server
	var orderQuery string
	offSet := filter.GetPage() * filter.GetPageSize()
	if strings.Trim(filter.GetSortBy(), " ") == "" {
		orderQuery = strings.Trim(filter.GetSortBy(), " ") + " " + TypeSortToString(filter.GetSort())
	}

	var result = u.db

	if filter.GetLimit() != -1 {
		result = result.Limit(int(filter.GetLimit()))
	}

	if filter.GetPage() != -1 && filter.GetPageSize() != -1 {
		result = result.Offset(int(offSet))
	}

	if filter.GetSortBy() != "" {
		result = result.Order(orderQuery)
	}

	result = result.Find(&server)

	if result.Error != nil {
		return nil, result.Error
	}
	return server, nil
}

func NewServerRepository(db *gorm.DB) *ServerRepositoryDecorator {
	return &ServerRepositoryDecorator{model.CreateServerRepository(db), db}
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

	if _, ok := result["Status"]; ok {
		if req.GetStatus() == pb.ServerStatus_OFF {
			result["Status"] = model.Off
		}
		if req.GetStatus() == pb.ServerStatus_ON {
			result["Status"] = model.On
		}
	}
	return result, nil
}
