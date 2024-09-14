package config

import (
	"log"

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

func (p *Project) GetDebug() bool {
	return p.Debug
}

func (p *Project) GetName() string {
	return p.Name
}

func (p *Project) GetEnvironment() string {
	return p.Environment
}

func (p *Project) GetVersion() string {
	return p.Version
}

func (p *Project) GetCommitHash() string {
	return p.CommitHash
}

// Grpc - contains parameter address grpc.
type Grpc struct {
	Host              string `yaml:"host" mapstructure:"host"`
	Port              int    `yaml:"port" mapstructure:"port"`
	MaxConnectionIdle int64  `yaml:"maxConnectionIdle" mapstructure:"maxConnectionIdle"`
	Timeout           int64  `yaml:"timeout" mapstructure:"timeout"`
	MaxConnectionAge  int64  `yaml:"maxConnectionAge" mapstructure:"maxConnectionAge"`
}

func (g *Grpc) GetGrpcHost() string {
	return g.Host
}

func (g *Grpc) GetGrpcPort() int {
	return g.Port
}

func (g *Grpc) GetGrpcMaxConnectionIdle() int64 {
	return g.MaxConnectionIdle
}

func (g *Grpc) GetGrpcTimeout() int64 {
	return g.Timeout
}

func (g *Grpc) GetMaxConnectionAge() int64 {
	return g.MaxConnectionAge
}

// Gateway - contains parameters for grpc-gateway port.
type Gateway struct {
	Host               string   `yaml:"host" mapstructure:"host"`
	Port               int      `yaml:"port" mapstructure:"port"`
	AllowedCORSOrigins []string `yaml:"allowedCorsOrigins" mapstructure:"allowedCorsOrigins"`
}

func (g *Gateway) GetGatewayHost() string {
	return g.Host
}

func (g *Gateway) GetGatewayPort() int {
	return g.Port
}

func (g *Gateway) GetGatewayAllowedCORSOrigins() []string {
	return g.AllowedCORSOrigins
}

// Swagger - contains parameters for swagger port.
type Swagger struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	GtAddr   string `yaml:"gtAddr" mapstructure:"gtAddr"`
	Filepath string `yaml:"filepath" mapstructure:"filepath"`
}

func (s *Swagger) GetSwaggerHost() string {
	return s.Host
}

func (s *Swagger) GetSwaggerPort() int {
	return s.Port
}

func (s *Swagger) GetGtAddr() string {
	return s.GtAddr
}

func (s *Swagger) GetFilepath() string {
	return s.Filepath
}

// Data - struct for data.
type Data struct {
	StockFilePath string `yaml:"stockFilePath" mapstructure:"stockFilePath"`
}

func (d *Data) GetStockFilePath() string {
	return d.StockFilePath
}

// Config - contains all configuration parameters in config package.
type Config struct {
	Project Project `yaml:"project" mapstructure:"project"`
	Grpc    Grpc    `yaml:"grpc" mapstructure:"grpc"`
	Gateway Gateway `yaml:"gateway" mapstructure:"gateway"`
	Swagger Swagger `yaml:"swagger" mapstructure:"swagger"`
	Data    Data    `yaml:"data" mapstructure:"data"`
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
		log.Printf("Error loading config file: %v", err)
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
		log.Printf("Error unmarshalling config: %v", err)
		return err
	}

	c.Project.Version = version
	c.Project.CommitHash = commitHash

	return nil
}

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

	// Data
	viper.SetDefault("data.stockFilePath", "data/stock-data.json")
}

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

		// Data
		"data.stockFilePath": "DATA_STOCK_FILEPATH",
	}

	for key, env := range envVars {
		if err := viper.BindEnv(key, env); err != nil {
			log.Printf("Error bind env: %v", err)
			return err
		}
	}
	return nil
}
