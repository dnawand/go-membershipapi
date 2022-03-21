package app

import (
	"errors"
	"time"

	"github.com/dnawand/go-subscriptionapi/pkg/domain"
)

type SubscriptionService struct {
	sr domain.SubscriptionRepository
	ur domain.UserRepository
	pr domain.ProductRepository
}

func NewSubscriptionService(sr domain.SubscriptionRepository, ur domain.UserRepository, pr domain.ProductRepository) *SubscriptionService {
	return &SubscriptionService{sr: sr, ur: ur, pr: pr}
}

func (ss *SubscriptionService) Subscribe(userID, productID, subscriptionPlanID string) (subscription domain.Subscription, err error) {
	var dataNotFoundErr *domain.DataNotFoundError

	user, err := ss.ur.Get(userID)
	if err != nil {
		if errors.As(err, &dataNotFoundErr) {
			return subscription, err
		}
		return subscription, domain.InternalError
	}

	if subscription, ok := getSubscription(user, productID); ok {
		return subscription, nil
	}

	now := time.Now()
	product, err := ss.pr.Get(productID)
	if err != nil {
		if errors.As(err, &dataNotFoundErr) {
			return domain.Subscription{}, err
		}
		return domain.Subscription{}, domain.InternalError
	}

	subscriptionPlan, ok := getSubscriptionPlan(subscriptionPlanID, product)
	if !ok {
		return subscription, &domain.DataNotFoundError{DataType: "subscription plan"}
	}

	subscription = domain.Subscription{
		ProductID:          product.ID,
		SubscriptionPlanID: subscriptionPlan.ID,
		StartDate:          now,
		EndDate:            addMonths(now, int(subscriptionPlan.Length)),
		PauseDate:          nil,
	}
	user = domain.User{
		ID:            userID,
		Subscriptions: []domain.Subscription{subscription},
	}

	subscription, err = ss.sr.Save(user)
	if err != nil {
		return domain.Subscription{}, domain.InternalError
	}

	subscription, err = ss.sr.Get(subscription.ID)
	if err != nil {
		return domain.Subscription{}, domain.InternalError
	}

	return subscription, nil
}

func (ss *SubscriptionService) Fetch(subscriptionID string) (domain.Subscription, error) {
	return ss.sr.Get(subscriptionID)
}

func (ss *SubscriptionService) List(userID string) ([]domain.Subscription, error) {
	return ss.sr.List(userID)
}

func (ss *SubscriptionService) Pause(subscriptionID string) (domain.Subscription, error) {
	panic("Implement me")
}

func (ss *SubscriptionService) Resume(subscriptionID string) (domain.Subscription, error) {
	panic("Implement me")
}

func (ss *SubscriptionService) Unsubscribe(subscriptionID string) (domain.Subscription, error) {
	panic("Implement me")
}

func getSubscriptionPlan(SubscriptionPlanID string, product domain.Product) (domain.SubscriptionPlan, bool) {
	for _, sp := range product.SubscriptionPlans {
		if sp.ID == SubscriptionPlanID {
			return sp, true
		}
	}

	return domain.SubscriptionPlan{}, false
}

func addMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, int(months), 0)
}

func getSubscription(user domain.User, productID string) (domain.Subscription, bool) {
	for _, s := range user.Subscriptions {
		if s.ProductID == productID {
			return s, true
		}
	}

	return domain.Subscription{}, false
}
