package domain

type UserService interface {
	Create(User) (User, error)
	Fetch(userID string) (User, error)
}

type ProductService interface {
	Create(Product) (Product, error)
	Fetch(productID string) (Product, error)
	List() ([]Product, error)
}

type SubscriptionService interface {
	Subscribe(userID, productID, productPlanID string, voucherID string) (Subscription, error)
	Fetch(userID, subscriptionID string) (Subscription, error)
	List(userID string) ([]Subscription, error)
	Pause(userID, subscriptionID string) (Subscription, error)
	Resume(userID, subscriptionID string) (Subscription, error)
	Unsubscribe(userID, subscriptionID string) (Subscription, error)
}

type DiscountService interface {
	ApplyDiscountOnPrice(price Money, v Voucher) (Money, error)
	ApplyDiscountOnTax(price Money, tax Money, v Voucher) (Money, error)
}
