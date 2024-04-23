package cache

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	gRedis "github.com/nguyentrunghieu15/vcs-be-prj/pkg/redis"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
)

type ServerCache struct {
	prefix string
	gedis  *gRedis.Gedis
}

func NewServerCache(config gRedis.GedisConfig) *ServerCache {
	return &ServerCache{
		prefix: "list",
		gedis:  gRedis.NewRedisClient(config),
	}
}

func (s *ServerCache) CheckExistsKey(key string) bool {
	result := s.gedis.GetAllKeyWithPattern(fmt.Sprintf("%v:hash:%v:serverId:*", s.prefix, key))
	return len(result) > 0
}

func (s *ServerCache) CacheListServerWithKey(key string, servers []model.Server) error {
	for _, server := range servers {
		result := s.gedis.Set(
			fmt.Sprintf("%v:hash:%v:serverId:%v", s.prefix, key, server.ID),
			s.convertServerToMapRedis(server))
		if result.Err() != nil {
			s.gedis.DelKeyWithPattern(fmt.Sprintf("%v:hash:%v:serverId:*", s.prefix, key))
			return result.Err()
		}
	}
	return nil
}

func (s *ServerCache) DelKey(key string) error {
	result := s.gedis.DelKeyWithPattern(fmt.Sprintf("%v:hash:%v:serverId:*", s.prefix, key))
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (s *ServerCache) DelKeyContainServerID(serverId string) error {
	result := s.gedis.DelKeyWithPattern(fmt.Sprintf("%v:hash:*:serverId:%v", s.prefix, serverId))
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func (s *ServerCache) UpdateOneServerForAllKey(server model.Server) error {
	allKeys := s.gedis.GetAllKeyWithPattern(
		fmt.Sprintf("%v:hash:*:serverId:%v", s.prefix, server.ID.String()))
	for _, v := range allKeys {
		result := s.gedis.Set(v, s.convertServerToMapRedis(server))
		if result.Err() != nil {
			return result.Err()
		}
	}
	return nil
}

func (s *ServerCache) GetListServerWithKey(key string) ([]model.Server, error) {
	result := make([]model.Server, 0)
	allKeys := s.gedis.GetAllKeyWithPattern(fmt.Sprintf("%v:hash:%v:serverId:*", s.prefix, key))
	for _, v := range allKeys {
		server := s.gedis.Get(v)
		if server.Err() != nil {
			return []model.Server{}, server.Err()
		}
		result = append(result, s.convertMapRedisToServerModel(server.Val()))
	}
	return result, nil
}

func (s *ServerCache) ClearCache() error {
	reuslt := s.gedis.FlushDb()
	return reuslt.Err()
}

func (s *ServerCache) convertServerToMapRedis(server model.Server) map[string]string {
	return map[string]string{
		"id":         server.ID.String(),
		"created_at": server.CreatedAt.Format(time.RFC3339),
		"updated_at": server.UpdatedAt.Format(time.RFC3339),
		"Deleted_at": server.DeletedAt.Time.Format(time.RFC3339),
		"created_by": strconv.Itoa(server.CreatedBy),
		"updated_by": strconv.Itoa(server.UpdatedBy),
		"deleted_by": strconv.Itoa(server.DeletedBy),
		"name":       server.Name,
		"ipv4":       server.Ipv4,
		"status":     string(server.Status),
	}
}

func parseStringToTime(t string) time.Time {
	result, _ := time.Parse(time.RFC3339, t)
	return result
}

func parseStringToInt(i string) int {
	result, _ := strconv.ParseInt(i, 2, 64)
	return int(result)
}

func (s *ServerCache) convertMapRedisToServerModel(server map[string]string) model.Server {
	return model.Server{
		ID:        uuid.MustParse(server["id"]),
		CreatedAt: parseStringToTime(server["created_at"]),
		UpdatedAt: parseStringToTime(server["created_at"]),
		DeletedAt: gorm.DeletedAt{Time: parseStringToTime(server["deleted_at"])},
		CreatedBy: parseStringToInt(server["created_by"]),
		UpdatedBy: parseStringToInt(server["updated_by"]),
		DeletedBy: parseStringToInt(server["deleted_by"]),
		Name:      server["name"],
		Status:    model.ServerStatus(server["status"]),
		Ipv4:      server["ipv4"],
	}
}
