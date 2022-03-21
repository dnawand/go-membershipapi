package repositories

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dnawand/go-subscriptionapi/pkg/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (ur *UserRepository) Save(user domain.User) (domain.User, error) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("error when generating id for user: %s\n", err.Error())
		return domain.User{}, fmt.Errorf("error when generating id for user: %w", err)
	}
	user.ID = newUUID.String()

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	ur.db.Create(&user)

	return user, nil
}

func (ur *UserRepository) Get(userID string) (domain.User, error) {
	var user domain.User

	tx := ur.db.
		Preload("Subscriptions.Product").
		Preload("Subscriptions.SubscriptionPlan").
		Find(&user, "id = ?", userID)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return domain.User{}, &domain.ErrDataNotFound{DataType: "user"}
		}
		return domain.User{}, fmt.Errorf("error when querying user: %w", tx.Error)
	}

	return user, nil
}
