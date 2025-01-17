package config

import (
	"context"
	"route256/utils/logger"

	"github.com/spf13/viper"
)

// Build information -ldflags.
var (
	version    string = "dev"
	commitHash string = "-"
)

// Project - contains all parameters project information.
type Project struct {
	Debug       bool   `yaml:"debug" mapstructure:"debug"`
	Name        string `yaml:"name" mapstructure:"name"`
	Environment string `yaml:"environment" mapstructure:"environment"`
	Version     string
	CommitHash  string
}

func (p *Project) GetDebug() bool         { return p.Debug }
func (p *Project) GetName() string        { return p.Name }
func (p *Project) GetEnvironment() string { return p.Environment }
func (p *Project) GetVersion() string     { return p.Version }
func (p *Project) GetCommitHash() string  { return p.CommitHash }

// Grpc - contains parameter address grpc.
type Grpc struct {
	Host              string `yaml:"host" mapstructure:"host"`
	Port              int    `yaml:"port" mapstructure:"port"`
	MaxConnectionIdle int64  `yaml:"maxConnectionIdle" mapstructure:"maxConnectionIdle"`
	Timeout           int64  `yaml:"timeout" mapstructure:"timeout"`
	MaxConnectionAge  int64  `yaml:"maxConnectionAge" mapstructure:"maxConnectionAge"`
}

func (g *Grpc) GetGrpcHost() string             { return g.Host }
func (g *Grpc) GetGrpcPort() int                { return g.Port }
func (g *Grpc) GetGrpcMaxConnectionIdle() int64 { return g.MaxConnectionIdle }
func (g *Grpc) GetGrpcTimeout() int64           { return g.Timeout }
func (g *Grpc) GetMaxConnectionAge() int64      { return g.MaxConnectionAge }

// Gateway - contains parameters for grpc-gateway port.
type Gateway struct {
	Host               string   `yaml:"host" mapstructure:"host"`
	Port               int      `yaml:"port" mapstructure:"port"`
	AllowedCORSOrigins []string `yaml:"allowedCorsOrigins" mapstructure:"allowedCorsOrigins"`
}

func (g *Gateway) GetGatewayHost() string                 { return g.Host }
func (g *Gateway) GetGatewayPort() int                    { return g.Port }
func (g *Gateway) GetGatewayAllowedCORSOrigins() []string { return g.AllowedCORSOrigins }

// Swagger - contains parameters for swagger port.
type Swagger struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	GtAddr   string `yaml:"gtAddr" mapstructure:"gtAddr"`
	Filepath string `yaml:"filepath" mapstructure:"filepath"`
	Dist     string `yaml:"dist" mapstructure:"dist"`
}

func (s *Swagger) GetSwaggerHost() string { return s.Host }
func (s *Swagger) GetSwaggerPort() int    { return s.Port }
func (s *Swagger) GetGtAddr() string      { return s.GtAddr }
func (s *Swagger) GetFilepath() string    { return s.Filepath }
func (s *Swagger) GetDist() string        { return s.Dist }

// Data - struct for data.
type Data struct {
	StockFilePath string `yaml:"stockFilePath" mapstructure:"stockFilePath"`
}

func (d *Data) GetStockFilePath() string { return d.StockFilePath }

// Database.
type Database struct {
	DSN              string   `yaml:"dsn"`
	Shards           []string `yaml:"shards" mapstructure:"shards"`
	ShardBucketCount int      `yaml:"shardBucketCount" mapstructure:"shardBucketCount"`
}

func (d *Database) GetDSN() string           { return d.DSN }
func (d *Database) GetShards() []string      { return d.Shards }
func (d *Database) GetShardBucketCount() int { return d.ShardBucketCount }

// Kafka.
type Kafka struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

func (k *Kafka) GetBrokers() []string { return k.Brokers }
func (k *Kafka) GetTopic() string     { return k.Topic }

// Jaeger - contains parameters for jaeger.
type Jaeger struct {
	URI string `yaml:"uri" mapstructure:"uri"`
}

func (j *Jaeger) GetURI() string { return j.URI }

// Metrics - contains parameters for metrics.
type Metrics struct {
	URI string `yaml:"uri" mapstructure:"uri"`
}

func (m *Metrics) GetURI() string { return m.URI }

// Graylog - contains parameters for graylog.
type Graylog struct {
	URI string `yaml:"uri" mapstructure:"uri"`
}

func (g *Graylog) GetURI() string { return g.URI }

// Config - contains all configuration parameters in config package.
type Config struct {
	Project  Project  `yaml:"project" mapstructure:"project"`
	Grpc     Grpc     `yaml:"grpc" mapstructure:"grpc"`
	Gateway  Gateway  `yaml:"gateway" mapstructure:"gateway"`
	Swagger  Swagger  `yaml:"swagger" mapstructure:"swagger"`
	Data     Data     `yaml:"data" mapstructure:"data"`
	Database Database `yaml:"database" mapstructure:"database"`
	Kafka    Kafka    `yaml:"kafka" mapstructure:"kafka"`
	Jaeger   Jaeger   `yaml:"jaeger" mapstructure:"jaeger"`
	Metrics  Metrics  `yaml:"metrics" mapstructure:"metrics"`
	Graylog  Graylog  `yaml:"graylog" mapstructure:"graylog"`
}

func NewConfig() *Config {
	return &Config{}
}

// ReadConfig - read configurations from default/file/env and init instance Config.
func (c *Config) ReadConfig(configPath string) error {

	// Set default values
	setDefaultValues()

	// Read config file
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		logger.Errorw(context.Background(), "Error loading config file", "error", err)
	}

	// Read env
	viper.AutomaticEnv()

	// Bind env variables
	if err := c.bindEnvVariables(); err != nil {
		return err
	}

	// Load config into struct
	err = viper.Unmarshal(&c)
	if err != nil {
		logger.Errorw(context.Background(), "Error unmarshalling config", "error", err)
		return err
	}

	c.Project.Version = version
	c.Project.CommitHash = commitHash

	return nil
}

// setDefaultValues function for set default values of config.
func setDefaultValues() {
	// Project
	viper.SetDefault("project.debug", "false")
	viper.SetDefault("project.name", "Cart")
	viper.SetDefault("project.environment", "development")

	// GRPC
	viper.SetDefault("grpc.host", "0.0.0.0")
	viper.SetDefault("grpc.port", 50051)
	viper.SetDefault("grpc.maxConnectionIdle", 5)
	viper.SetDefault("grpc.timeout", 15)
	viper.SetDefault("grpc.maxConnectionAge", 5)

	// Gateway
	viper.SetDefault("gateway.host", "0.0.0.0")
	viper.SetDefault("gateway.port", 50050)
	viper.SetDefault("gateway.allowedCorsOrigins", []string{
		"http://localhost:50052",
		"http://0.0.0.0:50052",
		"http://127.0.0.1:50052",
	})

	// Swagger
	viper.SetDefault("swagger.host", "0.0.0.0")
	viper.SetDefault("swagger.gtAddr", "0.0.0.0")
	viper.SetDefault("swagger.port", 50052)
	viper.SetDefault("swagger.filepath", "api/openapiv2/loms.swagger.json")
	viper.SetDefault("swagger.dist", "./swagger/dist")

	// Data
	viper.SetDefault("data.stockFilePath", "data/stock-data.json")

	// Database
	viper.SetDefault("database.dsn", "postgres://user:password@localhost:5432/postgres?sslmode=disable")
	viper.SetDefault("database.shards", []string{
		"postgres://user:password@localhost:5430/postgres?sslmode=disable",
		"postgres://user:password@localhost:5431/postgres?sslmode=disable",
	})
	viper.SetDefault("database.shardBucketCount", 1000)

	// Kafka
	viper.SetDefault("kafka.brokers", "localhost:9092")
	viper.SetDefault("kafka.topic", "loms.order-events")

	// Jaeger
	viper.SetDefault("jaeger.uri", "http://localhost:4318")

	// Metrics
	viper.SetDefault("metrics.uri", "http://localhost:2113")

	// Graylog
	viper.SetDefault("graylog.uri", "127.0.0.1:12201")
}

// bindEnvVariables function for bind env variables with config name.
func (c *Config) bindEnvVariables() error {
	envVars := map[string]string{
		// Project
		"project.debug":       "PROJECT_DEBUG",
		"project.name":        "PROJECT_NAME",
		"project.environment": "PROJECT_ENVIRONMENT",

		// gRPC
		"grpc.host":              "GRPC_HOST",
		"grpc.port":              "GRPC_PORT",
		"grpc.maxConnectionIdle": "GRPC_MAX_CONNECTION_IDLE",
		"grpc.timeout":           "GRPC_TIMEOUT",
		"grpc.maxConnectionAge":  "GRPC_MAX_CONNECTION_AGE",

		// Gateway
		"gateway.host":               "GATEWAY_HOST",
		"gateway.port":               "GATEWAY_PORT",
		"gateway.allowedCorsOrigins": "GATEWAY_ALLOWED_CORS_ORIGINS",

		// Swagger
		"swagger.host":     "SWAGGER_HOST",
		"swagger.port":     "SWAGGER_PORT",
		"swagger.gtAddr":   "SWAGGER_GT_ADDR",
		"swagger.filepath": "SWAGGER_FILEPATH",
		"swagger.dist":     "SWAGGER_DIST",

		// Data
		"data.stockFilePath": "DATA_STOCK_FILEPATH",

		// Database
		"database.dsn":              "DATABASE_DSN",
		"database.shards":           "DATABASE_SHARDS",
		"database.shardBucketCount": "DATABASE_SHARD_BUCKET_COUNT",

		// Kafka
		"kafka.brokers": "KAFKA_BROKERS",
		"kafka.topic":   "KAFKA_TOPIC",

		// Jaeger
		"jaeger.uri": "JAEGER_URI",

		// Metrics
		"metrics.uri": "METRICS_URI",

		// Graylog
		"graylog.uri": "GRAYLOG_URI",
	}

	for key, env := range envVars {
		if err := viper.BindEnv(key, env); err != nil {
			logger.Errorw(context.Background(), "Error bind env", "error", err)
			return err
		}
	}
	return nil
}
