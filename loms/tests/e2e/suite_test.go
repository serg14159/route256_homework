package suite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	loms_app "route256/loms/internal/app/loms"
	"route256/loms/internal/models"
	db "route256/loms/internal/pkg/db"
	kafka_producer "route256/loms/internal/pkg/kafka"
	mw "route256/loms/internal/pkg/mw"
	"route256/loms/internal/pkg/shard_manager"
	repo_order "route256/loms/internal/repository/orders"
	repo_outbox "route256/loms/internal/repository/outbox"
	repo_stock "route256/loms/internal/repository/stocks"
	loms_service "route256/loms/internal/service/loms"
	pb "route256/loms/pkg/api/loms/v1"
	"route256/loms/tests/e2e/migrations"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const (
	bufSize          = 1024 * 1024
	DSN              = "postgres://user:password@localhost:5432/postgres_test?sslmode=disable"
	ShardDSN1        = "postgres://user:password@localhost:5430/postgres_test?sslmode=disable"
	ShardDSN2        = "postgres://user:password@localhost:5431/postgres_test?sslmode=disable"
	ShardCount       = 2
	ShardBucketCount = 10
	KafkaBrokers     = "localhost:9092"
	KafkaTopic       = "test_topic"
)

// TSuite
type TSuite struct {
	suite.Suite
	conn          *grpc.ClientConn
	client        pb.LomsClient
	orderRepo     *repo_order.OrderRepository
	stockRepo     *repo_stock.StockRepository
	outboxRepo    *repo_outbox.OutboxRepository
	txManager     *db.TransactionManager
	grpcServer    *grpc.Server
	listener      *bufconn.Listener
	lomsApp       *loms_app.Service
	lomsService   *loms_service.LomsService
	kafkaProducer *kafka_producer.KafkaProducer
	kafkaConfig   *TestKafkaConfig
	dbPool        *pgxpool.Pool
	shardPools    []*pgxpool.Pool
	shardManager  *shard_manager.ShardManager
	gooseConns    []*sql.DB
}

// TestKafkaConfig
type TestKafkaConfig struct {
	brokers []string
	topic   string
}

// GetBrokers
func (c *TestKafkaConfig) GetBrokers() []string {
	return c.brokers
}

// GetTopic
func (c *TestKafkaConfig) GetTopic() string {
	return c.topic
}

// SetupSuite
func (s *TSuite) SetupSuite() {
	// Context
	ctx := context.Background()
	// Initialize bufconn listener
	s.listener = bufconn.Listen(bufSize)

	// DB connection
	dbConn, err := db.NewConnect(ctx, DSN)
	require.NoError(s.T(), err)
	s.dbPool = dbConn

	// DB connection for goose migrations
	s.gooseConns = make([]*sql.DB, 3)
	s.gooseConns[0], err = sql.Open("pgx", DSN)
	require.NoError(s.T(), err)
	s.gooseConns[1], err = sql.Open("pgx", ShardDSN1)
	require.NoError(s.T(), err)
	s.gooseConns[2], err = sql.Open("pgx", ShardDSN2)
	require.NoError(s.T(), err)

	// Run migrations up
	for i, conn := range s.gooseConns {
		if err := runMigrations(conn, "up"); err != nil {
			log.Fatalf("Failed to run migrations on shard %d: %v", i, err)
		}
	}

	// Initialize shard pools
	shards := []string{ShardDSN1, ShardDSN2}
	s.shardPools = make([]*pgxpool.Pool, len(shards))
	for i, dsn := range shards {
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			log.Fatalf("Failed to connect to shard %d: %v", i, err)
		}
		s.shardPools[i] = pool
	}

	// Initialize ShardManager
	shardFn := shard_manager.GetMurmur3ShardFn(len(s.shardPools))
	s.shardManager = shard_manager.NewShardManager(shardFn, s.shardPools, ShardBucketCount)

	// Initialize repositories
	s.orderRepo = repo_order.NewOrderRepository(s.shardManager)
	s.stockRepo = repo_stock.NewStockRepository(dbConn)
	s.outboxRepo = repo_outbox.NewOutboxRepository(dbConn)
	s.txManager = db.NewTransactionManager(dbConn)

	// Kafka configuration
	s.kafkaConfig = &TestKafkaConfig{
		brokers: []string{KafkaBrokers},
		topic:   KafkaTopic,
	}

	// Initialize Kafka producer
	kafkaProd, err := kafka_producer.NewKafkaProducer(s.kafkaConfig)
	require.NoError(s.T(), err)
	s.kafkaProducer = kafkaProd

	// Initialize LomsService
	lomsService := loms_service.NewService(s.orderRepo, s.stockRepo, s.outboxRepo, s.txManager, s.kafkaProducer)
	s.lomsService = lomsService
	s.lomsApp = loms_app.NewService(lomsService)

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
	pb.RegisterLomsServer(s.grpcServer, s.lomsApp)

	// Run GRPC server
	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Initialize GRPC client
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(s.bufDialer),
		grpc.WithInsecure(),
	)
	require.NoError(s.T(), err)

	s.conn = conn
	s.client = pb.NewLomsClient(conn)
}

// bufDialer returns connection via bufconn
func (s *TSuite) bufDialer(ctx context.Context, address string) (net.Conn, error) {
	return s.listener.Dial()
}

// TearDownSuite rolls back migrations and closes connections after all tests.
func (s *TSuite) TearDownSuite() {
	// Run migrations down
	for i, conn := range s.gooseConns {
		if err := runMigrations(conn, "down"); err != nil {
			log.Fatalf("Failed to run migrations on shard %d: %v", i, err)
		}
	}

	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.listener != nil {
		s.listener.Close()
	}
	if s.kafkaProducer != nil {
		s.kafkaProducer.Close()
	}
	if s.dbPool != nil {
		s.dbPool.Close()
	}
	for _, conn := range s.gooseConns {
		conn.Close()
	}
	for _, conn := range s.shardPools {
		conn.Close()
	}
}

// TearDownTest cleans up Kafka topic after each test.
func (s *TSuite) TearDownTest() {
	// Clean up Kafka topic
	admin, err := sarama.NewClusterAdmin(s.kafkaConfig.GetBrokers(), nil)
	require.NoError(s.T(), err)
	defer admin.Close()

	// Delete the topic
	err = admin.DeleteTopic(s.kafkaConfig.GetTopic())
	require.NoError(s.T(), err)

	// Recreate topic for next test
	err = admin.CreateTopic(s.kafkaConfig.GetTopic(), &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	require.NoError(s.T(), err)
}

// AfterTest is called after each test.
func (s *TSuite) AfterTest(suiteName, testName string) {
	s.TearDownTest()
}

// runMigrations
func runMigrations(db *sql.DB, action string) error {
	// Configure goose
	goose.SetBaseFS(migrations.EmbedFS)

	// Init version
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Error setting goose dialect: %v", err)
		return err
	}

	// Run migration
	var err error
	switch action {
	case "up":
		err = goose.Up(db, ".")
		if err != nil {
			log.Fatalf("Error applying migrations up: %v", err)
			return err
		}
		log.Println("Migrations successfully applied up.")
	case "down":
		err = goose.DownTo(db, ".", 0)
		if err != nil {
			log.Fatalf("Error rolling back migrations down: %v", err)
			return err
		}
		log.Println("Migrations successfully rolled back down.")
	default:
		log.Fatalf("Unknown migration action: %s. Use 'up' or 'down'.", action)
		return err
	}
	return nil
}

// consumeKafkaMessages consumes messages from Kafka topic.
func (s *TSuite) consumeKafkaMessages(_ context.Context, count int) ([]*sarama.ConsumerMessage, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewConsumer(s.kafkaConfig.GetBrokers(), config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	partitionConsumer, err := client.ConsumePartition(s.kafkaConfig.GetTopic(), 0, sarama.OffsetOldest)
	if err != nil {
		return nil, err
	}
	defer partitionConsumer.Close()

	messages := make([]*sarama.ConsumerMessage, 0, count)
	timeout := time.After(5 * time.Second)
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			messages = append(messages, msg)
			if len(messages) >= count {
				return messages, nil
			}
		case err := <-partitionConsumer.Errors():
			return nil, err
		case <-timeout:
			return messages, fmt.Errorf("timeout while waiting for messages")
		}
	}
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

	// Process outbox events
	err = s.lomsService.ProcessOutbox(ctx)
	require.NoError(s.T(), err)

	// Consume messages from Kafka
	messages, err := s.consumeKafkaMessages(ctx, 2)
	require.NoError(s.T(), err)
	require.Len(s.T(), messages, 2)

	// Verify messages
	for _, msg := range messages {
		var event models.OrderEvent
		err := json.Unmarshal(msg.Value, &event)
		require.NoError(s.T(), err)
		require.Equal(s.T(), orderID, event.OrderID)
	}

	// Check order status
	createdOrder, err := s.orderRepo.GetByID(ctx, orderID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), models.OrderStatusAwaitingPayment, createdOrder.Status)
	require.Equal(s.T(), models.UID(1), createdOrder.UserID)
	require.Len(s.T(), createdOrder.Items, 3)
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
		Status: models.OrderStatusAwaitingPayment,
	}

	// Create order in repository
	orderID, err := s.orderRepo.Create(ctx, order)
	require.NoError(s.T(), err)
	require.Greater(s.T(), orderID, models.OID(0))

	// Reserve items
	err = s.stockRepo.ReserveItems(ctx, nil, order.Items)
	require.NoError(s.T(), err)

	// Send OrderCancelRequest
	cancelReq := &pb.OrderCancelRequest{
		OrderID: int64(orderID),
	}

	cancelResp, err := s.client.OrderCancel(ctx, cancelReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), cancelResp)

	// Process outbox events
	err = s.lomsService.ProcessOutbox(ctx)
	require.NoError(s.T(), err)

	// Consume messages from Kafka
	messages, err := s.consumeKafkaMessages(ctx, 1)
	require.NoError(s.T(), err)
	require.Len(s.T(), messages, 1)

	// Verify message
	var event models.OrderEvent
	err = json.Unmarshal(messages[0].Value, &event)
	require.NoError(s.T(), err)
	require.Equal(s.T(), orderID, event.OrderID)
	require.Equal(s.T(), models.OrderStatusCancelled, event.Status)

	// Check order status
	updatedOrder, err := s.orderRepo.GetByID(ctx, orderID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), models.OrderStatusCancelled, updatedOrder.Status)
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
		Status: models.OrderStatusAwaitingPayment,
	}

	// Create order in repository
	orderID, err := s.orderRepo.Create(ctx, order)
	require.NoError(s.T(), err)
	require.Greater(s.T(), orderID, models.OID(0))

	// Reserve items
	err = s.stockRepo.ReserveItems(ctx, nil, order.Items)
	require.NoError(s.T(), err)

	// Send OrderPayRequest
	payReq := &pb.OrderPayRequest{
		OrderID: int64(orderID),
	}

	_, err = s.client.OrderPay(ctx, payReq)
	require.NoError(s.T(), err)

	// Process outbox events
	err = s.lomsService.ProcessOutbox(ctx)
	require.NoError(s.T(), err)

	// Consume messages from Kafka
	messages, err := s.consumeKafkaMessages(ctx, 1)
	require.NoError(s.T(), err)
	require.Len(s.T(), messages, 1)

	// Verify message
	var event models.OrderEvent
	err = json.Unmarshal(messages[0].Value, &event)
	require.NoError(s.T(), err)
	require.Equal(s.T(), orderID, event.OrderID)
	require.Equal(s.T(), models.OrderStatusPayed, event.Status)

	// Check order status
	updatedOrder, err := s.orderRepo.GetByID(ctx, orderID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), models.OrderStatusPayed, updatedOrder.Status)
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
	orderID, err := s.orderRepo.Create(ctx, order)
	require.NoError(s.T(), err)
	require.Greater(s.T(), orderID, models.OID(0))

	// Send OrderInfoRequest
	infoReq := &pb.OrderInfoRequest{
		OrderID: int64(orderID),
	}

	infoResp, err := s.client.OrderInfo(ctx, infoReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), infoResp)
	require.Equal(s.T(), string(models.OrderStatusNew), infoResp.Status)
	require.Equal(s.T(), int64(2), infoResp.User)
	require.Len(s.T(), infoResp.Items, 3)

	// Check order items
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
