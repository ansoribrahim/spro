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

func TestRepository_PostEstate(t *testing.T) {
	var (
		db   *sql.DB
		mock sqlmock.Sqlmock

		mockTime = time.Now()
		mockUUID = uuid.New()
		entity   = EstateEntity{
			Width:            5,
			Length:           10,
			TotalDistance:    0,
			TreeCount:        0,
			TreeMaxHeight:    0,
			TreeMinHeight:    0,
			TreeMedianHeight: 0,
			CreatedAt:        mockTime,
		}

		query = `INSERT INTO "estates" ("width","length","total_distance","tree_count","tree_max_height","tree_min_height","tree_median_height","created_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`
	)

	tests := []struct {
		name         string
		entity       EstateEntity
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
					WithArgs(entity.Width, entity.Length, entity.TotalDistance, entity.TreeCount, entity.TreeMaxHeight, entity.TreeMinHeight, entity.TreeMedianHeight, entity.CreatedAt).
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
		},
		{
			name:         "Insert Error",
			entity:       entity,
			expectedResp: nil,
			expectedErr:  sql.ErrNoRows,
			prepareMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(entity.Width, entity.Length, entity.TotalDistance, entity.TreeCount, entity.TreeMaxHeight, entity.TreeMinHeight, entity.TreeMedianHeight, entity.CreatedAt).
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

			resp, err := repo.PostEstate(context.Background(), tt.entity)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResp, resp)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
