package app

import (
	"errors"
	"time"

	"github.com/dnawand/go-membershipapi/internal/storage"
	"github.com/dnawand/go-membershipapi/pkg/domain"
	"github.com/dnawand/go-membershipapi/pkg/repositories"
)

const TrialPeriod = 1

type SubscriptionService struct {
	sr             domain.SubscriptionRepository
	ur             domain.UserRepository
	pr             domain.ProductRepository
	voucherStorage *storage.Store
	ds             domain.DiscountService
}

func NewSubscriptionService(
	sr domain.SubscriptionRepository,
	ur domain.UserRepository,
	pr domain.ProductRepository,
	vs *storage.Store,
	ds domain.DiscountService,
) *SubscriptionService {
	return &SubscriptionService{sr: sr, ur: ur, pr: pr, voucherStorage: vs, ds: ds}
}

func (ss *SubscriptionService) Subscribe(
	userID, productID, productPlanID string,
	voucherID string,
) (subscription domain.Subscription, err error) {
	subscription, err = ss.buildSubscription(userID, productID, productPlanID, voucherID)
	if err != nil {
		return subscription, err
	}
	if subscription.ID != "" {
		return subscription, nil
	}

	user := domain.User{
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

func (ss *SubscriptionService) buildSubscription(
	userID, productID, productPlanID, voucherID string,
) (subscription domain.Subscription, err error) {
	var voucher domain.Voucher

	if voucherID != "" {
		v, ok := ss.validateVoucher(voucherID)
		if !ok {
			return domain.Subscription{}, &domain.ErrInvalidArgument{Msg: "invalid voucher"}
		}
		voucher = v
	}

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

	product, err := ss.pr.Get(productID)
	if err != nil {
		if errors.As(err, &dataNotFoundErr) {
			return domain.Subscription{}, err
		}
		return domain.Subscription{}, domain.ErrInternal
	}

	productPlan, ok := getProductPlan(productPlanID, product)
	if !ok {
		return subscription, &domain.ErrDataNotFound{DataType: "product plan"}
	}

	price, err := ss.ds.ApplyDiscountOnPrice(productPlan.Price, voucher)
	if err != nil {
		return domain.Subscription{}, err
	}
	tax, err := ss.ds.ApplyDiscountOnTax(productPlan.Price, productPlan.Tax, voucher)
	if err != nil {
		return domain.Subscription{}, err
	}

	subscriptionPlan := domain.SubscriptionPlan{
		Plan: &domain.Plan{
			Length: productPlan.Length,
			Price:  price,
			Tax:    tax,
		},
		VoucherID: voucherID,
	}
	now := time.Now()
	trialDate := addMonths(now, TrialPeriod)
	startDate := trialDate.Add(1 * time.Hour)
	endDate := addMonths(startDate, productPlan.Length)
	subscription = domain.Subscription{
		ProductID:        product.ID,
		SubscriptionPlan: subscriptionPlan,
		TrialDate:        trialDate,
		StartDate:        startDate,
		EndDate:          &endDate,
		PauseDate:        nil,
		IsActive:         true,
	}

	return subscription, err
}

func (ss *SubscriptionService) validateVoucher(voucherID string) (domain.Voucher, bool) {
	if !ss.voucherStorage.Exist(voucherID) {
		return domain.Voucher{}, false
	}

	v, _ := ss.voucherStorage.Load(voucherID)
	voucher, ok := v.(domain.Voucher)
	if !ok {
		return domain.Voucher{}, false
	}

	if !voucher.IsActive {
		return domain.Voucher{}, false
	}

	return voucher, true
}

func getProductPlan(SubscriptionPlanID string, product domain.Product) (domain.ProductPlan, bool) {
	for _, p := range product.ProductPlans {
		if p.ID == SubscriptionPlanID {
			return p, true
		}
	}

	return domain.ProductPlan{}, false
}

func getSubscription(user domain.User, productID string) (domain.Subscription, bool) {
	for _, s := range user.Subscriptions {
		if s.ProductID == productID {
			return s, true
		}
	}

	return domain.Subscription{}, false
}

func addMonths(t time.Time, months int) time.Time {
	endDate := t.AddDate(0, months, 0)
	return endDate
}
