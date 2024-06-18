package repository

//
//func TestRepository_AdjustPlotForwardDistance(t *testing.T) {
//	var (
//		db        *sql.DB
//		mock      sqlmock.Sqlmock
//		mockUUID  = uuid.New()
//		estateID  = mockUUID
//		orderNum  = 5
//		addlDist  = 10
//		mockError = gorm.ErrInvalidData
//
//		query = `UPDATE "plots" SET "distance" = distance + ? WHERE estate_id = ? AND order_number > ?;`
//	)
//
//	tests := []struct {
//		name         string
//		estateID     uuid.UUID
//		currentOrder int
//		additional   int
//		expectedErr  error
//		prepareMock  func()
//	}{
//		{
//			name:         "Successful Update",
//			estateID:     estateID,
//			currentOrder: orderNum,
//			additional:   addlDist,
//			expectedErr:  nil,
//			prepareMock: func() {
//				mock.ExpectBegin()
//				mock.ExpectExec(regexp.QuoteMeta(query)).
//					WithoutArgs(addlDist, estateID, orderNum).
//					WillReturnResult(sqlmock.NewResult(0, 1))
//				mock.ExpectCommit()
//			},
//		},
//		{
//			name:         "Database Error",
//			estateID:     estateID,
//			currentOrder: orderNum,
//			additional:   addlDist,
//			expectedErr:  mockError,
//			prepareMock: func() {
//				mock.ExpectBegin()
//				mock.ExpectExec(regexp.QuoteMeta(query)).
//					WithArgs(addlDist, estateID, orderNum).
//					WillReturnError(mockError)
//				mock.ExpectRollback()
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// Initialize sqlmock and gorm.DB
//			db, mock, _ = sqlmock.New()
//			gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			// Set up mock expectations
//			tt.prepareMock()
//
//			// Initialize repository with mocked database connection
//			repo := NewRepository(NewRepositoryOptions{Db: gdb})
//
//			// Call the method under test
//			err = repo.AdjustPlotForwardDistance(context.Background(), tt.estateID, tt.currentOrder, tt.additional)
//
//			// Assert the expected error
//			assert.Equal(t, tt.expectedErr, err)
//
//			// Verify all expectations were met
//			assert.NoError(t, mock.ExpectationsWereMet())
//		})
//	}
//}
