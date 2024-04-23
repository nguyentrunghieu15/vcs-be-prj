package exporter

import (
	"fmt"
	"time"

	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
)

type ExporterServerRepo struct {
	db *gorm.DB
}

func NewExporterServerRepo(db *gorm.DB) *ExporterServerRepo {
	return &ExporterServerRepo{
		db: db,
	}
}

func (s *ExporterServerRepo) FindAllServer(req *pb.ExportServerRequest) ([]model.Server, error) {
	var servers []model.Server
	result := s.db

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
			case pb.FilterServer_ON:
				result = result.Where("status = ? ", model.On)
			case pb.FilterServer_OFF:
				result = result.Where("status = ? ", model.Off)
			}
		}

	}

	// Add pagination
	if pagination := req.GetPagination(); pagination != nil {
		if pagination.PageSize != nil {
			if pagination.FromPage != nil {
				result = result.Offset(int(pagination.GetFromPage()-1) * int(pagination.GetPageSize()))
			}
			if pagination.ToPage != nil {
				result = result.Limit(
					int(pagination.GetPageSize()) * int(pagination.GetToPage()-pagination.GetFromPage()),
				)
			}
			if orderBy := req.GetPagination().SortBy; orderBy != nil {
				if req.GetPagination().Sort != nil && req.GetPagination().GetSort() == pb.TypeSort_DESC {
					result = result.Order(fmt.Sprintf("%v %v", *orderBy, "DESC"))
				} else {
					result = result.Order(*orderBy)
				}
			}
		}
	}

	result = result.Find(&servers)
	if result.Error != nil {
		return nil, result.Error
	}
	return servers, nil
}
