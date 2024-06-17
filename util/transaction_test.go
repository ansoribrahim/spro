package util

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func handleTransaactionWithPanicTest(tx *gorm.DB, err error) {
	defer HandleTransaction(tx, err)
	panic(errors.New("err"))
}

func TestHandleTransaction(t *testing.T) {
	db, _, _ := sqlmock.New()
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.Nil(t, err)

	type args struct {
		tx  *gorm.DB
		err error
	}
	tests := []struct {
		name          string
		args          args
		isShouldPanic bool
	}{
		{
			name: "expect rollback",
			args: args{
				tx:  gdb,
				err: errors.New("text"),
			},
		},
		{
			name: "expect commit",
			args: args{
				tx: gdb,
			},
		},
		{
			name: "panic",
			args: args{
				tx: gdb,
			},
			isShouldPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isShouldPanic {
				assert.Panics(t, func() {
					handleTransaactionWithPanicTest(tt.args.tx, err)
				})
			} else {
				HandleTransaction(tt.args.tx, tt.args.err)
			}
		})
	}
}

func TestNewTxContext(t *testing.T) {
	ctx := context.TODO()
	tx := &gorm.DB{}
	type args struct {
		ctx context.Context
		tx  *gorm.DB
	}
	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{
			name: "normal",
			args: args{
				ctx: ctx,
				tx:  tx,
			},
			want: context.WithValue(ctx, "db", tx),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTxContext(tt.args.ctx, tt.args.tx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTxContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTxFromContext(t *testing.T) {
	ctx := context.TODO()
	db := &gorm.DB{}
	tx := &gorm.DB{}

	type args struct {
		ctx       context.Context
		defaultTx *gorm.DB
	}
	tests := []struct {
		name string
		args args
		want *gorm.DB
	}{
		{
			name: "normal without tx",
			args: args{
				ctx:       ctx,
				defaultTx: db,
			},
			want: db,
		},
		{
			name: "normal with tx",
			args: args{
				ctx:       context.WithValue(ctx, "db", tx),
				defaultTx: tx,
			},
			want: tx,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTxFromContext(tt.args.ctx, tt.args.defaultTx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTxFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
