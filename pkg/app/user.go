package app

import (
	"github.com/dnawand/go-subscriptionapi/pkg/domain"
)

type UserService struct {
	userRepository domain.UserRepository
}

func NewUserService(ur domain.UserRepository) *UserService {
	return &UserService{
		userRepository: ur,
	}
}

func (us *UserService) Create(user domain.User) (domain.User, error) {
	return us.userRepository.Save(user)
}
