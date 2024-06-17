package repository

import (
	"time"

	"github.com/google/uuid"
)

type EstateEntity struct {
	ID               uuid.UUID `gorm:"default:uuid_generate_v4()"`
	Width            int
	Length           int
	TotalDistance    int
	TreeCount        int
	TreeMaxHeight    int
	TreeMinHeight    int
	TreeMedianHeight int
	CreatedAt        time.Time
}

func (EstateEntity) TableName() string {
	return "estates"
}

// x & y max will be 50,000. it enough to use uint16 which could store until 65,535
type PlotEntity struct {
	ID          uuid.UUID `gorm:"default:uuid_generate_v4()"`
	EstateId    uuid.UUID
	X           uint16
	Y           uint16
	Distance    int
	OrderNumber int
	TreeHeight  int
	CreatedAt   time.Time
}

func (PlotEntity) TableName() string {
	return "plots"
}
