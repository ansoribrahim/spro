package service

import (
	_ "github.com/lib/pq"
	"gorm.io/gorm"

	"spgo/repository"
)

type Service struct {
	Repository repository.RepositoryInterface
	Db         *gorm.DB
}

type NewServiceOptions struct {
	Repository repository.RepositoryInterface
	Db         *gorm.DB
}

func NewService(opts NewServiceOptions) *Service {
	return &Service{
		Repository: opts.Repository,
		Db:         opts.Db,
	}
}
