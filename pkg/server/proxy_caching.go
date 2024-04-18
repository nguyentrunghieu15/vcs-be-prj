package server

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nguyentrunghieu15/vcs-be-prj/pkg/cache"
	gedis "github.com/nguyentrunghieu15/vcs-be-prj/pkg/redis"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
)

func HashMD5(value interface{}) (string, error) {
	serializedValue, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	h := md5.New()
	h.Write(serializedValue)
	hashValue := h.Sum(nil)
	return hex.EncodeToString(hashValue), nil
}

type ServerRepoProxy struct {
	serverPostgreRepo *ServerRepositoryDecorator
	redisCache        *cache.ServerCache
}

func NewServerRepoProxy(configRedis gedis.GedisConfig, db *gorm.DB) *ServerRepoProxy {
	return &ServerRepoProxy{
		serverPostgreRepo: NewServerRepository(db),
		redisCache:        cache.NewServerCache(configRedis),
	}
}

// CheckServerExists implements ServerRepo.
func (s *ServerRepoProxy) CheckServerExists(data map[string]interface{}) bool {
	return s.serverPostgreRepo.CheckServerExists(data)
}

// CountServers implements ServerRepo.
func (s *ServerRepoProxy) CountServers(
	query *string,
	filter *server.FilterServer) (int64, error) {
	return s.serverPostgreRepo.CountServers(query, filter)
}

// CreateBacth implements ServerRepo.
func (s *ServerRepoProxy) CreateBacth(
	userId uint64,
	data []map[string]interface{},
) (*server.ImportServerResponse, error) {
	return s.serverPostgreRepo.CreateBacth(userId, data)
}

// CreateServer implements ServerRepo.
func (s *ServerRepoProxy) CreateServer(
	server map[string]interface{}) (*model.Server, error) {
	return s.serverPostgreRepo.CreateServer(server)
}

// DeleteOneById implements ServerRepo.
func (s *ServerRepoProxy) DeleteOneById(serverId uuid.UUID) error {
	result := s.serverPostgreRepo.DeleteOneById(serverId)
	if result != nil {
		return result
	}
	result = s.redisCache.DelKeyContainServerID(serverId.String())
	return result
}

// DeleteOneByName implements ServerRepo.
func (s *ServerRepoProxy) DeleteOneByName(name string) error {
	server, _ := s.serverPostgreRepo.FindOneByName(name)
	result := s.serverPostgreRepo.DeleteOneByName(name)
	if result != nil {
		return result
	}
	result = s.redisCache.DelKeyContainServerID(server.ID.String())
	return result
}

// FindOneById implements ServerRepo.
func (s *ServerRepoProxy) FindOneById(id uuid.UUID) (*model.Server, error) {
	return s.serverPostgreRepo.FindOneById(id)
}

// FindOneByName implements ServerRepo.
func (s *ServerRepoProxy) FindOneByName(name string) (*model.Server, error) {
	return s.serverPostgreRepo.FindOneByName(name)
}

// FindServers implements ServerRepo.
func (s *ServerRepoProxy) FindServers(req *server.ListServerRequest) ([]model.Server, error) {
	if req.GetPagination() != nil && req.GetPagination().SortBy != nil &&
		req.GetPagination().GetSortBy() != "created_at" {
		return s.serverPostgreRepo.FindServers(req)
	}

	key, err := HashMD5(req)
	if err != nil {
		return []model.Server{}, err
	}

	checkExistsKey := s.redisCache.CheckExistsKey(key)
	if checkExistsKey {
		return s.redisCache.GetListServerWithKey(key)
	}

	result, err := s.serverPostgreRepo.FindServers(req)
	if err != nil {
		return result, err
	}

	err = s.redisCache.CacheListServerWithKey(key, result)
	return result, err
}

// UpdateOneById implements ServerRepo.
func (s *ServerRepoProxy) UpdateOneById(id uuid.UUID, data map[string]interface{}) (*model.Server, error) {
	result, err := s.serverPostgreRepo.UpdateOneById(id, data)
	if err != nil {
		return result, err
	}
	err = s.redisCache.UpdateOneServerForAllKey(*result)
	return result, err
}

// UpdateOneByName implements ServerRepo.
func (s *ServerRepoProxy) UpdateOneByName(name string, data map[string]interface{}) (*model.Server, error) {
	result, err := s.serverPostgreRepo.UpdateOneByName(name, data)
	if err != nil {
		return result, err
	}
	err = s.redisCache.UpdateOneServerForAllKey(*result)
	return result, err
}
