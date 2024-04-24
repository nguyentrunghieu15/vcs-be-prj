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

func AddPagination(statement *gorm.DB, req *user.ListUsersRequest) *gorm.DB {
	if req.Pagination != nil {
		if limit := req.Pagination.Limit; limit != nil && *limit >= 1 {
			statement = statement.Limit(int(*limit))
		}
		page := req.Pagination.Page
		pageSize := req.Pagination.PageSize
		if page != nil && pageSize != nil && *page > 0 && *pageSize > 0 {
			statement.Offset(int((*page - 1) * (*pageSize)))
		}
		if orderBy := req.Pagination.SortBy; orderBy != nil {
			if req.Pagination.Sort != nil && req.Pagination.Sort == server.TypeSort_DESC.Enum() {
				statement = statement.Order(fmt.Sprintf("%v %v", orderBy, "DESC"))
			} else {
				statement = statement.Order(orderBy)
			}

		}
	}
	return statement
}

func (u *UserRepositoryDecorator) FindUsers(req *user.ListUsersRequest) ([]model.User, error) {
	var user []model.User
	result := u.db
	AddPagination(result, req)
	result = result.Find(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}
