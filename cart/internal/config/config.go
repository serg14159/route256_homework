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

// Server - contains parameters for server address and port
type Server struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port string `yaml:"port" mapstructure:"port"`
}

func (s *Server) GetPort() string { return s.Port }
func (s *Server) GetHost() string { return s.Host }

// ProductService - contains parameters for ProductService.
type ProductService struct {
	ApiURI     string `yaml:"apiuri" mapstructure:"apiuri"`
	Token      string `yaml:"token" mapstructure:"token"`
	MaxRetries int    `yaml:"maxRetries" mapstructure:"maxRetries"`
}

func (ps *ProductService) GetURI() string     { return ps.ApiURI }
func (ps *ProductService) GetToken() string   { return ps.Token }
func (ps *ProductService) GetMaxRetries() int { return ps.MaxRetries }

// LomsService - contains parameters for server address and port
type LomsService struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port string `yaml:"port" mapstructure:"port"`
}

func (l *LomsService) GetPort() string { return l.Port }
func (l *LomsService) GetHost() string { return l.Host }

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

// Cache
type Cache struct {
	Capacity int `yaml:"capacity"`
}

func (c *Cache) GetCapacity() int { return c.Capacity }

// Redis
type Redis struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     string `yaml:"port" mapstructure:"port"`
	Password string `yaml:"password" mapstructure:"password"`
	DB       int    `yaml:"db" mapstructure:"db"`
	TTL      int    `yaml:"ttl" mapstructure:"ttl"`
}

func (r *Redis) GetHost() string     { return r.Host }
func (r *Redis) GetPort() string     { return r.Port }
func (r *Redis) GetPassword() string { return r.Password }
func (r *Redis) GetDB() int          { return r.DB }
func (r *Redis) GetTTL() int         { return r.TTL }

// Graylog - contains parameters for graylog.
type Graylog struct {
	URI string `yaml:"uri" mapstructure:"uri"`
}

func (g *Graylog) GetURI() string { return g.URI }

// Config - contains all configuration parameters in config package
type Config struct {
	Project        Project        `yaml:"project" mapstructure:"project"`
	Server         Server         `yaml:"server" mapstructure:"server"`
	ProductService ProductService `yaml:"productService" mapstructure:"productService"`
	LomsService    LomsService    `yaml:"lomsService" mapstructure:"lomsService"`
	Jaeger         Jaeger         `yaml:"jaeger" mapstructure:"jaeger"`
	Metrics        Metrics        `yaml:"metrics" mapstructure:"metrics"`
	Cache          Cache          `yaml:"cache" mapstructure:"cache"`
	Redis          Redis          `yaml:"redis" mapstructure:"redis"`
	Graylog        Graylog        `yaml:"graylog" mapstructure:"graylog"`
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

	// Server
	viper.SetDefault("server.port", "8082")
	viper.SetDefault("server.host", "localhost")

	// ProductService
	viper.SetDefault("productService.apiuri", "http://route256.pavl.uk:8080")
	viper.SetDefault("productService.token", "testtoken")
	viper.SetDefault("productService.maxRetries", "3")

	// LomsService
	viper.SetDefault("lomsService.host", "0.0.0.0")
	viper.SetDefault("lomsService.port", "50051")

	// Jaeger
	viper.SetDefault("jaeger.uri", "http://localhost:4318")

	// Metrics
	viper.SetDefault("metrics.uri", "http://localhost:2112")

	// Cache
	viper.SetDefault("cache.capacity", "100")

	// Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.ttl", 60)

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

		// Server
		"server.host": "SERVER_HOST",
		"server.port": "SERVER_PORT",

		// ProductService
		"productService.apiuri":     "PRODUCT_SERVICE_APIURI",
		"productService.token":      "PRODUCT_SERVICE_TOKEN",
		"productService.maxRetries": "PRODUCT_SERVICE_MAX_RETRIES",

		// LomsService
		"lomsService.host": "LOMS_SERVICE_HOST",
		"lomsService.port": "LOMS_SERVICE_PORT",

		// Jaeger
		"jaeger.uri": "JAEGER_URI",

		// Metrics
		"metrics.uri": "METRICS_URI",

		// Cache
		"cache.capacity": "CACHE_CAPACITY",

		// Redis
		"redis.host":     "REDIS_HOST",
		"redis.port":     "REDIS_PORT",
		"redis.password": "REDIS_PASSWORD",
		"redis.db":       "REDIS_DB",
		"redis.ttl":      "REDIS_TTL",

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
