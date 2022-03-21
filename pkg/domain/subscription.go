package domain

import (
	"time"

	"gorm.io/gorm"
)

type Subscription struct {
	ID                 string           `json:"id" gorm:"type:uuid"`
	Product            Product          `json:"product"`
	ProductID          string           `json:"-" gorm:"type:uuid"`
	SubscriptionPlan   SubscriptionPlan `json:"subscriptionPlan"`
	SubscriptionPlanID string           `json:"-"`
	StartDate          time.Time        `json:"startDate"`
	EndDate            time.Time        `json:"endDate"`
	PauseDate          *time.Time       `json:"pauseDate,omitempty"`
	IsPaused           bool             `json:"paused"`
	IsActive           bool             `json:"-"`
	UserID             string           `json:"-" gorm:"type:uuid"`
	CreatedAt          time.Time        `json:"-"`
	UpdatedAt          time.Time        `json:"-"`
	DeletedAt          gorm.DeletedAt   `json:"-" gorm:"index"`
}
