package repository

//
//func TestRepository_GetEstate(t *testing.T) {
//	var (
//		db       *sql.DB
//		mock     sqlmock.Sqlmock
//		mockUUID = uuid.New()
//		mockID   = mockUUID
//		mockData = EstateEntity{
//			ID:               mockID,
//			Width:            10,
//			Length:           20,
//			TotalDistance:    500,
//			TreeCount:        50,
//			TreeMaxHeight:    15,
//			TreeMinHeight:    5,
//			TreeMedianHeight: 10,
//			CreatedAt:        time.Now(),
//		}
//	)
//
//	tests := []struct {
//		name         string
//		id           uuid.UUID
//		expectedData EstateEntity
//		expectedErr  error
//		prepareMock  func()
//	}{
//		{
//			name:         "Successful Retrieval",
//			id:           mockID,
//			expectedData: mockData,
//			expectedErr:  nil,
//			prepareMock: func() {
//				rows := sqlmock.NewRows([]string{
//					"id", "width", "length", "total_distance", "tree_count",
//					"tree_max_height", "tree_min_height", "tree_median_height", "created_at",
//				}).AddRow(
//					mockData.ID, mockData.Width, mockData.Length, mockData.TotalDistance,
//					mockData.TreeCount, mockData.TreeMaxHeight, mockData.TreeMinHeight,
//					mockData.TreeMedianHeight, mockData.CreatedAt,
//				)
//
//				// Use regexp.QuoteMeta on the SQL query template
//				queryTemplate := `SELECT * FROM "estates" WHERE id = ? ORDER BY "estates"."id" LIMIT 1`
//				queryPattern := regexp.QuoteMeta(queryTemplate)
//
//				mock.ExpectQuery(queryPattern).
//					WithArgs(mockID).
//					WillReturnRows(rows)
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			db, mock, _ = sqlmock.New()
//			gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			tt.prepareMock()
//
//			repo := NewRepository(NewRepositoryOptions{Db: gdb})
//
//			estate, err := repo.GetEstate(context.Background(), tt.id)
//
//			assert.Equal(t, tt.expectedErr, err)
//			assert.Equal(t, tt.expectedData.ID, estate.ID)
//			assert.Equal(t, tt.expectedData.Width, estate.Width)
//			assert.Equal(t, tt.expectedData.Length, estate.Length)
//			assert.Equal(t, tt.expectedData.TotalDistance, estate.TotalDistance)
//			assert.Equal(t, tt.expectedData.TreeCount, estate.TreeCount)
//			assert.Equal(t, tt.expectedData.TreeMaxHeight, estate.TreeMaxHeight)
//			assert.Equal(t, tt.expectedData.TreeMinHeight, estate.TreeMinHeight)
//			assert.Equal(t, tt.expectedData.TreeMedianHeight, estate.TreeMedianHeight)
//
//			// Compare time using Round to avoid precision issues
//			assert.WithinDuration(t, tt.expectedData.CreatedAt.Round(time.Second), estate.CreatedAt.Round(time.Second), time.Second)
//
//			assert.NoError(t, mock.ExpectationsWereMet())
//		})
//	}
//}
