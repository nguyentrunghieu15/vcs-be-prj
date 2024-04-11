package server

import (
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
)

type ServerRepositoryDecorator struct {
	*model.ServerRepository
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) *ServerRepositoryDecorator {
	return &ServerRepositoryDecorator{model.CreateServerRepository(db), db}
}
