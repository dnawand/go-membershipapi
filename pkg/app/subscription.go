package app

import (
	"errors"
	"time"

	"github.com/dnawand/go-subscriptionapi/pkg/domain"
	"github.com/dnawand/go-subscriptionapi/pkg/repositories"
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
	var dataNotFoundErr *domain.ErrDataNotFound

	user, err := ss.ur.Get(userID)
	if err != nil {
		if errors.As(err, &dataNotFoundErr) {
			return subscription, err
		}
		return subscription, domain.ErrInternal
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
		return domain.Subscription{}, domain.ErrInternal
	}

	subscriptionPlan, ok := getSubscriptionPlan(subscriptionPlanID, product)
	if !ok {
		return subscription, &domain.ErrDataNotFound{DataType: "subscription plan"}
	}

	subscription = domain.Subscription{
		ProductID:          product.ID,
		SubscriptionPlanID: subscriptionPlan.ID,
		StartDate:          now,
		EndDate:            addMonths(now, subscriptionPlan.Length),
		PauseDate:          nil,
		IsActive:           true,
	}
	user = domain.User{
		ID:            userID,
		Subscriptions: []domain.Subscription{subscription},
	}

	subscription, err = ss.sr.Save(user)
	if err != nil {
		return domain.Subscription{}, domain.ErrInternal
	}

	subscription, err = ss.sr.Get(subscription.ID)
	if err != nil {
		return domain.Subscription{}, domain.ErrInternal
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
	var dataNotFoundErr *domain.ErrDataNotFound

	subscription, err := ss.sr.Get(subscriptionID)
	if err != nil {
		if errors.As(err, &dataNotFoundErr) {
			return subscription, err
		}
		return subscription, domain.ErrInternal
	}

	if !subscription.IsActive {
		return subscription, domain.ErrForbidden
	}

	if subscription.IsPaused {
		return subscription, nil
	}

	now := time.Now()
	subscription.PauseDate = &now
	subscription.EndDate = nil
	subscription.IsPaused = true

	toUpdate := domain.ToUpdate{
		repositories.PauseDate: subscription.PauseDate,
		repositories.EndDate:   subscription.EndDate,
		repositories.IsPaused:  subscription.IsPaused,
	}

	subscription, err = ss.sr.Update(subscription, toUpdate)
	if err != nil {
		if errors.As(err, &dataNotFoundErr) {
			return subscription, err
		}
		return subscription, domain.ErrForbidden
	}

	return subscription, nil
}

func (ss *SubscriptionService) Resume(subscriptionID string) (domain.Subscription, error) {
	var dataNotFoundErr *domain.ErrDataNotFound

	subscription, err := ss.sr.Get(subscriptionID)
	if err != nil {
		if errors.As(err, &dataNotFoundErr) {
			return subscription, err
		}
		return subscription, domain.ErrForbidden
	}

	if !subscription.IsActive {
		return subscription, domain.ErrForbidden
	}

	if !subscription.IsPaused {
		return subscription, nil
	}

	previousEndDate := addMonths(subscription.StartDate, subscription.SubscriptionPlan.Length)
	diff := previousEndDate.Sub(*subscription.PauseDate)
	newEndDate := time.Now().Add(diff)

	subscription.PauseDate = nil
	subscription.EndDate = &newEndDate
	subscription.IsPaused = false

	toUpdate := domain.ToUpdate{
		repositories.PauseDate: subscription.PauseDate,
		repositories.EndDate:   subscription.EndDate,
		repositories.IsPaused:  subscription.IsPaused,
	}

	subscription, err = ss.sr.Update(subscription, toUpdate)
	if err != nil {
		if errors.As(err, &dataNotFoundErr) {
			return subscription, err
		}
		return subscription, domain.ErrInternal
	}

	return subscription, nil
}

func (ss *SubscriptionService) Unsubscribe(subscriptionID string) (domain.Subscription, error) {
	var dataNotFoundErr *domain.ErrDataNotFound

	subscription, err := ss.sr.Get(subscriptionID)
	if err != nil {
		if errors.As(err, &dataNotFoundErr) {
			return subscription, err
		}
		return subscription, domain.ErrInternal
	}

	if !subscription.IsActive {
		return subscription, nil
	}

	subscription.IsActive = false

	toUpdate := domain.ToUpdate{
		repositories.IsActive: subscription.IsActive,
	}

	subscription, err = ss.sr.Update(subscription, toUpdate)
	if err != nil {
		if errors.As(err, &dataNotFoundErr) {
			return subscription, err
		}
		return subscription, domain.ErrInternal
	}

	return subscription, nil
}

func getSubscriptionPlan(SubscriptionPlanID string, product domain.Product) (domain.SubscriptionPlan, bool) {
	for _, sp := range product.SubscriptionPlans {
		if sp.ID == SubscriptionPlanID {
			return sp, true
		}
	}

	return domain.SubscriptionPlan{}, false
}

func getSubscription(user domain.User, productID string) (domain.Subscription, bool) {
	for _, s := range user.Subscriptions {
		if s.ProductID == productID {
			return s, true
		}
	}

	return domain.Subscription{}, false
}

func addMonths(t time.Time, months int) *time.Time {
	endDate := t.AddDate(0, months, 0)
	return &endDate
}
