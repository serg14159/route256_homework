package service_test

import (
	"errors"
	service "route256/cart/internal/service/cart"
	"route256/cart/internal/service/cart/mock"

	"testing"

	"github.com/gojuno/minimock/v3"
	"go.uber.org/goleak"
)

var ErrRepository = errors.New("repository error")

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// setup function for setup initializes the mocks and the CartService for the tests.
func setup(t *testing.T) (*mock.ICartRepositoryMock, *mock.IProductServiceMock, *mock.ILomsServiceMock, *service.CartService) {
	ctrl := minimock.NewController(t)

	// Create mocks for ICartRepository and IProductService
	repoMock := mock.NewICartRepositoryMock(ctrl)
	productServiceMock := mock.NewIProductServiceMock(ctrl)
	lomsServiceMock := mock.NewILomsServiceMock(ctrl)
	// Initialize the service with the mocks
	service := service.NewService(repoMock, productServiceMock, lomsServiceMock)

	return repoMock, productServiceMock, lomsServiceMock, service
}
