package util

import (
	"context"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func HandleTransaction(tx *gorm.DB, err error) {
	if p := recover(); p != nil {
		tx.Rollback()
		logrus.Panic(p)
	} else if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
}

func NewTxContext(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, "db", tx)
}

func GetTxFromContext(ctx context.Context, defaultTx *gorm.DB) *gorm.DB {
	txVal := ctx.Value("db")
	tx, ok := txVal.(*gorm.DB)
	if !ok {
		return defaultTx
	}
	return tx
}
