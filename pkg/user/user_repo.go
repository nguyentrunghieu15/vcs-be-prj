package user

import (
	"fmt"

	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
)

type UserRepositoryDecorator struct {
	model.IUserRepository
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepositoryDecorator {
	return &UserRepositoryDecorator{model.NewUserRepository(db), db}
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

func (u *UserRepositoryDecorator) FindUsers(req *user.ListUsersRequest) ([]model.User, error) {
	var user []model.User
	result := u.db
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
			if req.GetPagination().Sort != nil && req.GetPagination().Sort == server.TypeSort_DESC.Enum() {
				result = result.Order(fmt.Sprintf("%v %v", orderBy, "DESC"))
			} else {
				result = result.Order(orderBy)
			}

		}
	}
	result = result.Find(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}
