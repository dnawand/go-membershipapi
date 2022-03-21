package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            string         `json:"id" gorm:"type:uuid"`
	Name          string         `json:"name"`
	Email         string         `json:"email" gorm:"uniqueIndex"`
	Subscriptions []Subscription `json:"subscriptions,omitempty"`
	CreatedAt     time.Time      `json:"-"`
	UpdatedAt     time.Time      `json:"-"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate is a Gorm hook interface.
// func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
// 	tx.Statement.SetColumn("ID", u.ID)
// 	return
// }
