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
	Subscribe(userID, productID, subscriptionPlanID string) (Subscription, error)
	Fetch(subscriptionID string) (Subscription, error)
	List(userID string) ([]Subscription, error)
	Pause(subscriptionID string) (Subscription, error)
	Resume(subscriptionID string) (Subscription, error)
	Unsubscribe(subscriptionID string) (Subscription, error)
}
