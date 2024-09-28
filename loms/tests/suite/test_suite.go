package suite

import (
	"context"
	"log"
	"net"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"route256/loms/internal/app/loms"
	"route256/loms/internal/config"
	"route256/loms/internal/models"
	db "route256/loms/internal/pkg/db"
	mw "route256/loms/internal/pkg/mv"
	repo_order "route256/loms/internal/repository/orders"
	repo_stock "route256/loms/internal/repository/stocks"
	service_loms "route256/loms/internal/service/loms"
	pb "route256/loms/pkg/api/loms/v1"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

const bufSize = 1024 * 1024

// TSuite
type TSuite struct {
	suite.Suite
	conn        *grpc.ClientConn
	client      pb.LomsClient
	orderRepo   *repo_order.OrderRepository
	stockRepo   *repo_stock.StockRepository
	txManager   *db.TransactionManager
	grpcServer  *grpc.Server
	listener    *bufconn.Listener
	lomsService *loms.Service
}

const DSN = "postgres://user:password@localhost:5434/postgres?sslmode=disable"

// SetupSuite
func (s *TSuite) SetupSuite() {
	// Init bufconn listener
	s.listener = bufconn.Listen(bufSize)

	// DB connection
	dbConn, err := db.NewConnect(context.Background(), &config.Database{
		DSN: DSN,
	})
	require.NoError(s.T(), err)

	// Init repository
	s.orderRepo = repo_order.NewOrderRepository(dbConn)
	s.stockRepo = repo_stock.NewStockRepository(dbConn)
	s.txManager = db.NewTransactionManager(dbConn)

	// Init LomsService
	lomsServiceImpl := service_loms.NewService(s.orderRepo, s.stockRepo, s.txManager)
	s.lomsService = loms.NewService(lomsServiceImpl)

	// Create GRPC server
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				mw.Logger,
				mw.Validate,
				grpcrecovery.UnaryServerInterceptor(),
			),
		),
	)

	// Register LomsService on GRPC server
	pb.RegisterLomsServer(s.grpcServer, s.lomsService)

	// Run GRPC server
	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			log.Fatalf("GRPC server failed: %v", err)
		}
	}()

	// Init GRPC connect with bufconn
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(s.bufDialer),
		grpc.WithInsecure(),
	)
	require.NoError(s.T(), err)

	s.conn = conn
	s.client = pb.NewLomsClient(conn)
}

// bufDialer returns a connection via bufconn
func (s *TSuite) bufDialer(ctx context.Context, address string) (net.Conn, error) {
	return s.listener.Dial()
}

// TearDownSuite
func (s *TSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.listener != nil {
		s.listener.Close()
	}
}

// TestOrderCancel_Success
func (s *TSuite) TestOrderCancel_Success() {
	ctx := context.Background()

	// Create order
	order := models.Order{
		UserID: 1,
		Items: []models.Item{
			{SKU: 1076963, Count: 2},
			{SKU: 1148162, Count: 3},
			{SKU: 1002, Count: 1},
		},
		Status: models.OrderStatusNew,
	}

	// Create order in repository
	orderID, err := s.orderRepo.Create(ctx, nil, order)
	require.NoError(s.T(), err)
	require.Greater(s.T(), orderID, models.OID(0))

	// ReserveItems
	err = s.stockRepo.ReserveItems(ctx, nil, order.Items)
	require.NoError(s.T(), err)

	// Send OrderCancelRequest
	cancelReq := &pb.OrderCancelRequest{
		OrderID: int64(orderID),
	}

	cancelResp, err := s.client.OrderCancel(ctx, cancelReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cancelResp)

	// Check order status
	updatedOrder, err := s.orderRepo.GetByID(ctx, nil, orderID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), models.OrderStatusCancelled, updatedOrder.Status)
}

// TestOrderCreate_Success
func (s *TSuite) TestOrderCreate_Success() {
	ctx := context.Background()

	// Create OrderCreateRequest
	createReq := &pb.OrderCreateRequest{
		User: 1,
		Items: []*pb.Item{
			{Sku: 1076963, Count: 2},
			{Sku: 1148162, Count: 3},
			{Sku: 1002, Count: 1},
		},
	}

	// Send OrderCreateRequest
	res, err := s.client.OrderCreate(ctx, createReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), res)
	require.Greater(s.T(), res.OrderID, int64(0))

	orderID := models.OID(res.OrderID)

	// Check order status
	createdOrder, err := s.orderRepo.GetByID(ctx, nil, orderID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), models.OrderStatusAwaitingPayment, createdOrder.Status)
	require.Equal(s.T(), models.UID(1), createdOrder.UserID)
	require.Len(s.T(), createdOrder.Items, 3)
}

// TestOrderInfo_Success
func (s *TSuite) TestOrderInfo_Success() {
	ctx := context.Background()

	// Create order
	order := models.Order{
		UserID: 2,
		Items: []models.Item{
			{SKU: 1076963, Count: 2},
			{SKU: 1148162, Count: 3},
			{SKU: 1002, Count: 1},
		},
		Status: models.OrderStatusNew,
	}

	// Create order in repository
	orderID, err := s.orderRepo.Create(ctx, nil, order)
	require.NoError(s.T(), err)
	require.Greater(s.T(), orderID, models.OID(0))

	// Send OrderInfoRequest
	infoReq := &pb.OrderInfoRequest{
		OrderID: int64(orderID),
	}

	infoResp, err := s.client.OrderInfo(ctx, infoReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), infoResp)
	require.Equal(s.T(), "new", infoResp.Status)
	require.Equal(s.T(), int64(2), infoResp.User)
	require.Len(s.T(), infoResp.Items, 3)

	// Check order info
	for _, item := range infoResp.Items {
		switch item.Sku {
		case 1076963:
			require.Equal(s.T(), uint32(2), item.Count)
		case 1148162:
			require.Equal(s.T(), uint32(3), item.Count)
		case 1002:
			require.Equal(s.T(), uint32(1), item.Count)
		default:
			require.Fail(s.T(), "Unexpected SKU in order items")
		}
	}
}

// TestOrderPay_Success
func (s *TSuite) TestOrderPay_Success() {
	ctx := context.Background()

	// Create order
	order := models.Order{
		UserID: 3,
		Items: []models.Item{
			{SKU: 1005, Count: 5},
		},
		Status: models.OrderStatusNew,
	}

	// Create order in repository
	orderID, err := s.orderRepo.Create(ctx, nil, order)
	require.NoError(s.T(), err)
	require.Greater(s.T(), orderID, models.OID(0))

	// Reserve
	err = s.stockRepo.ReserveItems(ctx, nil, order.Items)
	require.NoError(s.T(), err)

	// Set status AwaitingPayment
	err = s.orderRepo.SetStatus(ctx, nil, orderID, models.OrderStatusAwaitingPayment)
	require.NoError(s.T(), err)

	// Send OrderPayRequest
	payReq := &pb.OrderPayRequest{
		OrderID: int64(orderID),
	}

	_, err = s.client.OrderPay(ctx, payReq)
	require.NoError(s.T(), err)

	// Check status
	updatedOrder, err := s.orderRepo.GetByID(ctx, nil, orderID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), models.OrderStatusPayed, updatedOrder.Status)
}

// TestStocksInfo_Success
func (s *TSuite) TestStocksInfo_Success() {
	ctx := context.Background()

	SKU := uint32(1004)

	// Send StocksInfoRequest
	stocksInfoReq := &pb.StocksInfoRequest{
		Sku: SKU,
	}

	stocksInfoResp, err := s.client.StocksInfo(ctx, stocksInfoReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), stocksInfoResp)
	require.Equal(s.T(), uint64(260), stocksInfoResp.Count)
}
