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
				result = result.Order(fmt.Sprintf("%v %v", *orderBy, "DESC"))
			} else {
				result = result.Order(*orderBy)
			}
		}
	}

	result = result.Find(&servers)
	if result.Error != nil {
		return nil, result.Error
	}
	return servers, nil
}

func (s *ServerRepositoryDecorator) GetAllServers() ([]model.Server, error) {
	var servers []model.Server
	err := s.db.Find(&servers)
	return servers, err.Error
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

func (s *ServerRepositoryDecorator) CheckServerExists(data map[string]interface{}) bool {
	var count int64
	result := s.db.Model(&model.Server{})
	if id, ok := data["id"]; ok {
		result = result.Where("id = ?", id)
		if name, ok := data["name"]; ok {
			result = result.Or("name = ?", name)
		}
	} else {
		if name, ok := data["name"]; ok {
			result = result.Where("name = ?", name)
		}
	}

	result = result.Count(&count)
	if result.Error != nil || count > 0 {
		return true
	}
	return false
}

func ConvertServerModelMapToServerProto(server map[string]interface{}) *pb.Server {
	// Convert the map to JSON
	jsonData, _ := json.Marshal(server)
	var structData model.Server
	json.Unmarshal(jsonData, &structData)
	return ConvertServerModelToServerProto(structData)
}

func ConvertListServerModelMapToListServerProto(s []map[string]interface{}) []*server.Server {
	var result []*server.Server = make([]*server.Server, 0)
	for _, v := range s {
		result = append(result, ConvertServerModelMapToServerProto(v))
	}
	return result
}

func (s *ServerRepositoryDecorator) CreateBacth(
	userId uint64,
	data []map[string]interface{},
) (*pb.ImportServerResponse, error) {
	importServer := make([]map[string]interface{}, 0)
	resImportServer := make([]map[string]interface{}, 0)
	abortServer := make([]map[string]interface{}, 0)
	for _, v := range data {
		if s.CheckServerExists(v) {
			abortServer = append(abortServer, v)
		} else {
			v["created_at"] = time.Now()
			v["created_by"] = userId
			importServer = append(importServer, v)
			resImportServer = append(resImportServer, v)
		}
	}
	if len(importServer) != 0 {
		result := s.db.Model(&model.Server{}).Create(&importServer)
		if result.Error != nil {
			return nil, result.Error
		}
	}
	return &pb.ImportServerResponse{
		NumServerImported: int64(len(resImportServer)),
		ServerImported:    ConvertListServerModelMapToListServerProto(resImportServer),
		NumServerFail:     int64(len(abortServer)),
		ServerFail:        ConvertListServerModelMapToListServerProto(abortServer),
	}, nil
}
