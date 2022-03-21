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
