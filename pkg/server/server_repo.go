package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
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

func (s *ServerRepositoryDecorator) CountServers(query *string, filter *server.FilterServer) (int64, error) {
	result := s.db.Model(&model.Server{})

	//Add query
	if query != nil {
		result = result.Where("name LIKE ?", "%"+*query+"%").Or("ipv4 LIKE ?", "%"+*query+"%")
	}

	//Add filter
	if filter != nil {
		if createAtFrom := filter.CreatedAtFrom; createAtFrom != nil {
			convertedTime, err := time.Parse(time.RFC3339, *createAtFrom)
			if err != nil {
				return 0, err
			}
			result = result.Where("created_at > ?", convertedTime)
		}
		if createdAtTo := filter.CreatedAtTo; createdAtTo != nil {
			convertedTime, err := time.Parse(time.RFC3339, *createdAtTo)
			if err != nil {
				return 0, err
			}
			result = result.Where("created_at < ?", convertedTime)
		}
		if updatedAtFrom := filter.UpdatedAtFrom; updatedAtFrom != nil {
			convertedTime, err := time.Parse(time.RFC3339, *updatedAtFrom)
			if err != nil {
				return 0, err
			}
			result = result.Where("updated_at > ?", convertedTime)
		}
		if updatedAtTo := filter.UpdatedAtTo; updatedAtTo != nil {
			convertedTime, err := time.Parse(time.RFC3339, *updatedAtTo)
			if err != nil {
				return 0, err
			}
			result = result.Where("updated_at < ?", convertedTime)
		}

		if status := filter.Status; status != nil {
			switch *status {
			case pb.ServerStatus_ON:
				result = result.Where("status = ? ", model.On)
			case pb.ServerStatus_OFF:
				result = result.Where("status = ? ", model.Off)
			}
		}

	}
	var count int64
	result = result.Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return count, nil
}

func (s *ServerRepositoryDecorator) FindServers(req *server.ListServerRequest) ([]model.Server, error) {
	var servers []model.Server
	result := s.db

	//Add query
	if query := req.Query; query != nil {
		result = result.Where("name LIKE ?", "%"+*query+"%").Or("ipv4 LIKE ?", "%"+*query+"%")
	}

	//Add filter
	if filter := req.GetFilter(); filter != nil {
		if createAtFrom := filter.CreatedAtFrom; createAtFrom != nil {
			convertedTime, err := time.Parse(time.RFC3339, *createAtFrom)
			if err != nil {
				return nil, err
			}
			result = result.Where("created_at > ?", convertedTime)
		}
		if createdAtTo := filter.CreatedAtTo; createdAtTo != nil {
			convertedTime, err := time.Parse(time.RFC3339, *createdAtTo)
			if err != nil {
				return nil, err
			}
			result = result.Where("created_at < ?", convertedTime)
		}
		if updatedAtFrom := filter.UpdatedAtFrom; updatedAtFrom != nil {
			convertedTime, err := time.Parse(time.RFC3339, *updatedAtFrom)
			if err != nil {
				return nil, err
			}
			result = result.Where("updated_at > ?", convertedTime)
		}
		if updatedAtTo := filter.UpdatedAtTo; updatedAtTo != nil {
			convertedTime, err := time.Parse(time.RFC3339, *updatedAtTo)
			if err != nil {
				return nil, err
			}
			result = result.Where("updated_at < ?", convertedTime)
		}

		if status := filter.Status; status != nil {
			switch *status {
			case pb.ServerStatus_ON:
				result = result.Where("status = ? ", model.On)
			case pb.ServerStatus_OFF:
				result = result.Where("status = ? ", model.Off)
			}
		}

	}

	// Add pagination
	if req.GetPagination() != nil {
		if limit := req.GetPagination().Limit; limit != nil && *limit > 1 {
			result = result.Limit(int(*limit))
		}
		page := req.GetPagination().Page
		pageSize := req.GetPagination().PageSize
		if page != nil && pageSize != nil && *page > 0 && *pageSize > 0 {
			result.Offset(int((*page - 1) * (*pageSize)))
		}
		if orderBy := req.GetPagination().SortBy; orderBy != nil {
			if req.GetPagination().Sort != nil && req.GetPagination().Sort == pb.TypeSort_DESC.Enum() {
				result = result.Order(fmt.Sprintf("%v %v", orderBy, "DESC"))
			} else {
				result = result.Order(orderBy)
			}
		}
	}

	result = result.Find(&servers)
	if result.Error != nil {
		return nil, result.Error
	}
	return servers, nil
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
