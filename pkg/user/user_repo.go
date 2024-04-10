package user

import (
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
)

type UserRepositoryDecorator struct {
	*model.UserRepository
}

func NewUserRepository(db *gorm.DB) *UserRepositoryDecorator {
	return &UserRepositoryDecorator{model.CreateUserRepository(db)}
}

func (u *UserRepositoryDecorator) FindUsers() {

}
