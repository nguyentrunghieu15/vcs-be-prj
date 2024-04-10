package user

import (
	"strings"

	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
)

type UserRepositoryDecorator struct {
	*model.UserRepository
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepositoryDecorator {
	return &UserRepositoryDecorator{model.CreateUserRepository(db), db}
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

func (u *UserRepositoryDecorator) FindUsers(filter model.FilterQueryInterface) ([]model.User, error) {
	var user []model.User
	var orderQuery string
	offSet := filter.GetPage() * filter.GetPageSize()
	if strings.Trim(filter.GetSortBy(), " ") == "" {
		orderQuery = strings.Trim(filter.GetSortBy(), " ") + " " + TypeSortToString(filter.GetSort())
	}

	result := u.db.Limit(int(filter.GetLimit())).
		Offset(int(offSet)).
		Order(orderQuery).
		Find(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}
