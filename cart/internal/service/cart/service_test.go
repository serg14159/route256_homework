package service

import (
	"errors"
	"route256/cart/internal/service/cart/mock"

	"testing"

	"github.com/gojuno/minimock/v3"
)

var ErrRepository = errors.New("repository error")

// Function for setup initializes the mocks and the CartService for the tests.
func setup(t *testing.T) (*mock.ICartRepositoryMock, *mock.IProductServiceMock, *CartService) {
	ctrl := minimock.NewController(t)

	// Create mocks for ICartRepository and IProductService
	repoMock := mock.NewICartRepositoryMock(ctrl)
	productServiceMock := mock.NewIProductServiceMock(ctrl)

	// Initialize the service with the mocks
	service := NewService(repoMock, productServiceMock)

	return repoMock, productServiceMock, service
}
