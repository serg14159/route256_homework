package service_test

import (
	"errors"
	service "route256/loms/internal/service/loms"
	"route256/loms/internal/service/loms/mock"

	"testing"

	"github.com/gojuno/minimock/v3"
)

var ErrRepository = errors.New("repository error")

// setup function for setup initializes the mocks and the CartService for the tests.
func setup(t *testing.T) (*mock.IOrderRepositoryMock, *mock.IStockRepositoryMock, *mock.IOutboxRepositoryMock, *mock.IProducerMock, *mock.ITxManagerMock, *service.LomsService) {
	ctrl := minimock.NewController(t)

	// Create mocks for ICartRepository and IProductService
	orderRepoMock := mock.NewIOrderRepositoryMock(ctrl)
	stockRepoMock := mock.NewIStockRepositoryMock(ctrl)
	outboxRepoMock := mock.NewIOutboxRepositoryMock(ctrl)
	txManagerMock := mock.NewITxManagerMock(ctrl)
	producerMock := mock.NewIProducerMock(ctrl)

	// Initialize the service with the mocks
	service := service.NewService(orderRepoMock, stockRepoMock, outboxRepoMock, txManagerMock, producerMock)

	return orderRepoMock, stockRepoMock, outboxRepoMock, producerMock, txManagerMock, service
}
