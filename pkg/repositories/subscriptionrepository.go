package repositories

import (
	"errors"
	"fmt"
	"time"

	"github.com/dnawand/go-membershipapi/internal/storage"
	"github.com/dnawand/go-membershipapi/pkg/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	EndDate   domain.Column = "end_date"
	PauseDate domain.Column = "pause_date"
	IsPaused  domain.Column = "is_paused"
	IsActive  domain.Column = "is_active"
)

type SubscriptionRepository struct {
	db             *gorm.DB
	voucherStorage *storage.Store
}

func NewSubscriptionRepository(db *gorm.DB, vs *storage.Store) *SubscriptionRepository {
	return &SubscriptionRepository{
		db:             db,
		voucherStorage: vs,
	}
}

func (sr *SubscriptionRepository) Save(user domain.User) (domain.Subscription, error) {
	if len(user.Subscriptions) != 1 {
		return domain.Subscription{}, &domain.ErrInvalidArgument{Msg: "user must have at least one subscription"}
	}

	now := time.Now()
	subscriptionIndex := 0
	subscriptionID, subscriptionPlanID, err := generateIDs()
	if err != nil {
		return domain.Subscription{}, domain.ErrInternal
	}

	user.Subscriptions[subscriptionIndex].ID = subscriptionID
	user.Subscriptions[subscriptionIndex].CreatedAt = now
	user.Subscriptions[subscriptionIndex].UpdatedAt = now
	user.Subscriptions[subscriptionIndex].SubscriptionPlan.ID = subscriptionPlanID
	productSubscription := user.Subscriptions[subscriptionIndex]

	err = sr.db.Transaction(func(tx *gorm.DB) error {
		txErr := tx.Model(&user).Association("Subscriptions").Append(&productSubscription)
		if txErr != nil {
			return txErr
		}

		txErr = tx.Model(&productSubscription).Association("SubscriptionPlan").Append(&productSubscription.SubscriptionPlan)
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("error when saving subscription: %w", err)
	}

	return user.Subscriptions[subscriptionIndex], nil
}

func (sr *SubscriptionRepository) Get(subscriptionID string) (domain.Subscription, error) {
	var subscription domain.Subscription

	tx := sr.db.
		Preload("Product").
		Preload("SubscriptionPlan").
		Find(&subscription, "id = ?", subscriptionID)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return domain.Subscription{}, &domain.ErrDataNotFound{DataType: "product"}
		}
		return domain.Subscription{}, fmt.Errorf("error when getting subscription from db: %w", tx.Error)
	}

	data, _ := sr.voucherStorage.Load(subscription.SubscriptionPlan.VoucherID)
	voucher, _ := data.(domain.Voucher)
	subscription.SubscriptionPlan.Voucher = voucher

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
			return subscriptions, &domain.ErrDataNotFound{DataType: "subscription list"}
		}
		return nil, fmt.Errorf("error when querying subscriptions: %w", tx.Error)
	}

	return subscriptions, nil
}

func (sr *SubscriptionRepository) Update(subscription domain.Subscription, updates domain.ToUpdate) (domain.Subscription, error) {
	colAndVal := map[string]interface{}{}

	for k, v := range updates {
		colAndVal[string(k)] = v
	}

	if tx := sr.db.Model(&subscription).Select("*").Updates(colAndVal); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return domain.Subscription{}, &domain.ErrDataNotFound{DataType: "subscription"}
		}
		return domain.Subscription{}, fmt.Errorf("error when updating subscription: %w", tx.Error)
	}

	return subscription, nil
}

func generateIDs() (string, string, error) {
	subscriptionUUID, err := uuid.NewRandom()
	if err != nil {
		return "", "", fmt.Errorf("error when generating id for subscription: %w", err)
	}

	subscriptionPlanUUID, err := uuid.NewRandom()
	if err != nil {
		return "", "", fmt.Errorf("error when generating id for subscription plan: %w", err)
	}

	return subscriptionUUID.String(), subscriptionPlanUUID.String(), nil
}
