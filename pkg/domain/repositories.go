package domain

type UserRepository interface {
	Save(User) (User, error)
}

type ProductRepository interface {
	Save(Product) (Product, error)
	Get(productID string) (Product, error)
	List() ([]Product, error)
}

type SubscriptionRepository interface {
	Save(Subscription) Subscription
	Get(subscriptionID string) Subscription
	Update(Subscription) Subscription
}
