// This file contains the repository implementation layer.
package repository

import (
	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

type Repository struct {
	Db *gorm.DB
}

type NewRepositoryOptions struct {
	Db *gorm.DB
}

func NewRepository(opts NewRepositoryOptions) *Repository {
	return &Repository{
		Db: opts.Db,
	}
}
