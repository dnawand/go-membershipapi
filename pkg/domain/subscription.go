package domain

import (
	"time"

	"gorm.io/gorm"
)

type Subscription struct {
	ID        string         `json:"id" gorm:"type:uuid"`
	Product   Product        `json:"product"`
	StartDate time.Time      `json:"startDate"`
	EndDate   time.Time      `json:"endDate"`
	PauseDate time.Time      `json:"pauseDate"`
	IsPaused  bool           `json:"paused"`
	IsActive  bool           `json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	ProductID string         `json:"-" gorm:"type:uuid"`
	UserID    string         `json:"-" gorm:"type:uuid"`
}

// BeforeCreate is a Gorm hook interface.
func (s *Subscription) BeforeCreate(tx *gorm.DB) (err error) {
	tx.Statement.SetColumn("ID", s.ID)
	return
}
