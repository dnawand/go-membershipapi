package domain

import (
	"time"

	"gorm.io/gorm"
)

type VoucherType string

const (
	VoucherFixedAmount VoucherType = "FixedAmount"
	VoucherPercentage  VoucherType = "Percentage"
)

type Product struct {
	ID           string         `json:"id" gorm:"type:uuid;uniqueIndex"`
	Name         string         `json:"name"`
	ProductPlans []ProductPlan  `json:"plans,omitempty"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type Subscription struct {
	ID               string           `json:"id" gorm:"type:uuid;uniqueIndex"`
	Product          Product          `json:"product"`
	ProductID        string           `json:"-" gorm:"type:uuid"`
	SubscriptionPlan SubscriptionPlan `json:"plan"`
	TrialDate        time.Time        `json:"trialDate"`
	StartDate        time.Time        `json:"startDate"`
	EndDate          *time.Time       `json:"endDate,omitempty"`
	PauseDate        *time.Time       `json:"pauseDate,omitempty"`
	IsPaused         bool             `json:"paused"`
	IsActive         bool             `json:"active"`
	UserID           string           `json:"-" gorm:"type:uuid"`
	CreatedAt        time.Time        `json:"-"`
	UpdatedAt        time.Time        `json:"-"`
	DeletedAt        gorm.DeletedAt   `json:"-" gorm:"index"`
}

type Plan struct {
	ID        string         `json:"id" gorm:"type:uuid;uniqueIndex"`
	Length    int            `json:"length"`
	Price     Money          `json:"price" gorm:"type:string"`
	Tax       Money          `json:"tax" gorm:"type:string"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type ProductPlan struct {
	*Plan
	ProductID string `json:"-" gorm:"type:uuid"`
}

type SubscriptionPlan struct {
	*Plan
	Voucher        Voucher `json:"voucher" gorm:"-:all"`
	VoucherID      string  `json:"-" gorm:"type:uuid"`
	SubscriptionID string  `json:"-" gorm:"type:uuid"`
}

type Voucher struct {
	ID       string      `json:"number"`
	Type     VoucherType `json:"-"`
	Discount string      `json:"-"`
	IsActive bool        `json:"-"`
}
