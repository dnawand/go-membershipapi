package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/dnawand/go-subscriptionapi/pkg/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

func (sr *SubscriptionRepository) Save(user domain.User) (domain.Subscription, error) {
	if len(user.Subscriptions) != 1 {
		return domain.Subscription{}, errors.New("user must have at least one subscription")
	}

	now := time.Now()
	subscriptionIndex := 0
	subscriptionID, err := uuid.NewRandom()
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("error when generating id for user: %w", err)
	}

	user.Subscriptions[subscriptionIndex].ID = subscriptionID.String()
	user.Subscriptions[subscriptionIndex].CreatedAt = now
	user.Subscriptions[subscriptionIndex].UpdatedAt = now
	pSubscription := user.Subscriptions[subscriptionIndex]

	err = sr.db.Model(&user).Association("Subscriptions").Append(&pSubscription)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("could not save subscription: %w", err)
	}

	return user.Subscriptions[subscriptionIndex], nil
}

func (sr *SubscriptionRepository) Get(subscriptionID string) (domain.Subscription, error) {
	var subscription domain.Subscription

	tx := sr.db.Debug().
		Preload("Product").
		Preload("SubscriptionPlan").
		Find(&subscription, "id = ?", subscriptionID)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return domain.Subscription{}, &domain.DataNotFoundError{DataType: "product"}
		}
	}

	return subscription, nil
}

func (sr *SubscriptionRepository) List(userID string) ([]domain.Subscription, error) {
	var subscriptions = []domain.Subscription{}

	tx := sr.db.
		Preload("Product").
		Preload("SubscriptionPlan").
		Joins("right join users on users.id = subscriptions.user_id").
		Where("user_id = ?", userID).
		Find(&subscriptions)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) || len(subscriptions) == 0 {
			return subscriptions, &domain.DataNotFoundError{DataType: "subscription list"}
		}

		return nil, fmt.Errorf("error when querying subscriptions: %w", tx.Error)
	}

	return subscriptions, nil
}

func (sr *SubscriptionRepository) Update(subscription domain.Subscription) (domain.Subscription, error) {
	panic("Implement me")
}
