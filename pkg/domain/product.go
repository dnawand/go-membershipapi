package domain

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID                string             `json:"id" gorm:"type:uuid"`
	Name              string             `json:"name"`
	SubscriptionPlans []SubscriptionPlan `json:"subscriptionPlans,omitempty"`
	CreatedAt         time.Time          `json:"-"`
	UpdatedAt         time.Time          `json:"-"`
	DeletedAt         gorm.DeletedAt     `json:"-" gorm:"index"`
}

type SubscriptionPlan struct {
	ID        string         `json:"id" gorm:"type:uuid"`
	Length    int32          `json:"length"`
	Price     Money          `json:"price" gorm:"type:string"`
	Tax       Money          `json:"tax" gorm:"type:string"`
	ProductID string         `json:"-" gorm:"type:uuid"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
