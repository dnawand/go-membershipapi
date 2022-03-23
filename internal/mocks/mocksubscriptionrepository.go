package mocks

import "github.com/dnawand/go-membershipapi/pkg/domain"

type MockSubscriptionRepository struct {
	SaveFunc   func(domain.User) (domain.Subscription, error)
	GetFunc    func(subscriptionID string) (domain.Subscription, error)
	ListFunc   func(userID string) ([]domain.Subscription, error)
	UpdateFunc func(domain.Subscription, domain.ToUpdate) (domain.Subscription, error)
}

func (msr *MockSubscriptionRepository) Save(u domain.User) (domain.Subscription, error) {
	return msr.SaveFunc(u)
}

func (msr *MockSubscriptionRepository) Get(subscriptionID string) (domain.Subscription, error) {
	return msr.GetFunc(subscriptionID)
}

func (msr *MockSubscriptionRepository) List(userID string) ([]domain.Subscription, error) {
	return msr.ListFunc(userID)
}

func (msr *MockSubscriptionRepository) Update(s domain.Subscription, toUpdate domain.ToUpdate) (domain.Subscription, error) {
	return msr.UpdateFunc(s, toUpdate)
}
