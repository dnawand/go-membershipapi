package domain

type UserService interface {
	Create(User) (User, error)
}

type ProductService interface {
	Create(Product) (Product, error)
	Fetch(productID string) (Product, error)
	List() ([]Product, error)
}

type SubscriptionService interface {
	Subscribe(User, Product) (Subscription, error)
	Fetch(subscriptionID string) (Subscription, error)
	List(userID string) ([]Subscription, error)
	Pause(subscriptionID string) (Subscription, error)
	Resume(subscriptionID string) (Subscription, error)
	Unsubscribe(subscriptionID string) (Subscription, error)
}
