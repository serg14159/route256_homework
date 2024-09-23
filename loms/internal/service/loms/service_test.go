package service

import (
	"errors"
	"route256/loms/internal/service/loms/mock"

	"testing"

	"github.com/gojuno/minimock/v3"
)

var ErrRepository = errors.New("repository error")

// Function for setup initializes the mocks and the CartService for the tests.
func setup(t *testing.T) (*mock.IOrderRepositoryMock, *mock.IStockRepositoryMock, *LomsService) {
	ctrl := minimock.NewController(t)

	// Create mocks for ICartRepository and IProductService
	orderRepoMock := mock.NewIOrderRepositoryMock(ctrl)
	stockRepoMock := mock.NewIStockRepositoryMock(ctrl)

	// Initialize the service with the mocks
	service := NewService(orderRepoMock, stockRepoMock)

	return orderRepoMock, stockRepoMock, service
}
