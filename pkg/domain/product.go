package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// CurrencyCode values are represented by ISO 4217 codes.
type CurrencyCode string

const (
	CurrencyEUR CurrencyCode = "EUR"
)

type Money struct {
	Code   string `json:"code"`
	Number string `json:"number"`
}

type Product struct {
	ID                string             `json:"id" gorm:"type:uuid"`
	Name              string             `json:"name"`
	SubscriptionPlans []SubscriptionPlan `json:"subscriptionPlans"`
	CreatedAt         time.Time          `json:"-"`
	UpdatedAt         time.Time          `json:"-"`
	DeletedAt         gorm.DeletedAt     `json:"-" gorm:"index"`
}

type SubscriptionPlan struct {
	ID        string         `json:"-" gorm:"type:uuid"`
	Length    int32          `json:"length"`
	Price     Money          `json:"price" gorm:"type:string"`
	Tax       Money          `json:"tax" gorm:"type:string"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	ProductID string         `json:"-" gorm:"type:uuid"`
}

// // BeforeCreate is a Gorm hook interface.
// func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
// 	tx.Statement.SetColumn("ID", p.ID)
// 	return
// }

// // BeforeCreate is a Gorm hook interface.
// func (sp *SubscriptionPlan) BeforeCreate(tx *gorm.DB) (err error) {
// 	tx.Statement.SetColumn("ID", sp.ID)
// 	return
// }

func (m *Money) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		log.Println("could not convert value from db into string")
		return fmt.Errorf("could not convert value from db into string")
	}

	money := Money{}
	err := json.Unmarshal([]byte(str), &money)
	if err != nil {
		log.Println("could not json into Money")
		return fmt.Errorf("could not json into Money")
	}

	*m = money

	return nil
}

func (m Money) Value() (driver.Value, error) {
	json, err := json.Marshal(m)
	if err != nil {
		log.Println("could not convert Money into json")
		return nil, fmt.Errorf("could not convert Money into json")
	}

	return json, nil
}
