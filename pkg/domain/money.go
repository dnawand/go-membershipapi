package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
)

// CurrencyCode values are represented by ISO 4217 codes.
type CurrencyCode string

const (
	CurrencyEUR CurrencyCode = "EUR"
)

type Money struct {
	Code   CurrencyCode `json:"code"`
	Number string       `json:"number"`
}

func (m *Money) Scan(value interface{}) error {
	b, ok := value.([]byte) // SQLite stores string as bytes
	if !ok {
		log.Println("could not convert value from db into bytes")
		return fmt.Errorf("could not convert value from db into bytes")
	}

	money := Money{}
	err := json.Unmarshal(b, &money)
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
