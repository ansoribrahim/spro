package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRepository_PostPlot(t *testing.T) {
	var (
		db   *sql.DB
		mock sqlmock.Sqlmock

		mockTime = time.Now()
		mockUUID = uuid.New()
		entity   = PlotEntity{
			EstateId:    mockUUID,
			X:           5,
			Y:           10,
			Distance:    34,
			OrderNumber: 1,
			TreeHeight:  5,
			CreatedAt:   mockTime,
		}

		query = `INSERT INTO "plots" ("estate_id","x","y","distance","order_number","tree_height","created_at") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`
	)

	tests := []struct {
		name         string
		entity       PlotEntity
		expectedResp *uuid.UUID
		expectedErr  error
		prepareMock  func()
	}{
		{
			name:         "Successful Insert",
			entity:       entity,
			expectedResp: &mockUUID,
			expectedErr:  nil,
			prepareMock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(mockUUID)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(entity.EstateId, entity.X, entity.Y, entity.Distance, entity.OrderNumber, entity.TreeHeight, entity.CreatedAt).
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
		},
		{
			name:         "Insert Error",
			entity:       entity,
			expectedResp: nil,
			expectedErr:  sql.ErrNoRows, // Example error, adjust as per your application's error handling
			prepareMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(entity.EstateId, entity.X, entity.Y, entity.Distance, entity.OrderNumber, entity.TreeHeight, entity.CreatedAt).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, _ = sqlmock.New()
			gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}

			tt.prepareMock()

			repo := NewRepository(NewRepositoryOptions{Db: gdb})

			resp, err := repo.PostPlot(context.Background(), tt.entity)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
