package domain

type UserRepository interface {
	Save(User) (User, error)
	Get(userID string) (User, error)
}

type ProductRepository interface {
	Save(Product) (Product, error)
	Get(productID string) (Product, error)
	List() ([]Product, error)
}

type SubscriptionRepository interface {
	Save(User) (Subscription, error)
	Get(subscriptionID string) (Subscription, error)
	List(userID string) ([]Subscription, error)
	Update(Subscription) (Subscription, error)
}
