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

// Server - contains parameters for server address and port
type Server struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port string `yaml:"port" mapstructure:"port"`
}

func (s *Server) GetPort() string {
	return s.Port
}

func (s *Server) GetHost() string {
	return s.Host
}

// ProductService - contains parameters for ProductService.
type ProductService struct {
	ApiURI     string `yaml:"apiuri" mapstructure:"apiuri"`
	Token      string `yaml:"token" mapstructure:"token"`
	MaxRetries int    `yaml:"maxRetries" mapstructure:"maxRetries"`
}

func (ps *ProductService) GetURI() string {
	return ps.ApiURI
}

func (ps *ProductService) GetToken() string {
	return ps.Token
}

func (ps *ProductService) GetMaxRetries() int {
	return ps.MaxRetries
}

// LomsService - contains parameters for server address and port
type LomsService struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port string `yaml:"port" mapstructure:"port"`
}

func (l *LomsService) GetPort() string {
	return l.Port
}

func (l *LomsService) GetHost() string {
	return l.Host
}

// Config - contains all configuration parameters in config package
type Config struct {
	Project        Project        `yaml:"project" mapstructure:"project"`
	Server         Server         `yaml:"server" mapstructure:"server"`
	ProductService ProductService `yaml:"productService" mapstructure:"productService"`
	LomsService    LomsService    `yaml:"lomsService" mapstructure:"lomsService"`
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
}

// bindEnvVariables function for bind env variables with config name.
func (c *Config) bindEnvVariables() error {
	envVars := map[string]string{
		"project.debug":             "PROJECT_DEBUG",
		"project.name":              "PROJECT_NAME",
		"project.environment":       "PROJECT_ENVIRONMENT",
		"server.host":               "SERVER_HOST",
		"server.port":               "SERVER_PORT",
		"productService.apiuri":     "PRODUCT_SERVICE_APIURI",
		"productService.token":      "PRODUCT_SERVICE_TOKEN",
		"productService.maxRetries": "PRODUCT_SERVICE_MAX_RETRIES",
		"lomsService.host":          "LOMS_SERVICE_HOST",
		"lomsService.port":          "LOMS_SERVICE_PORT",
	}

	for key, env := range envVars {
		if err := viper.BindEnv(key, env); err != nil {
			log.Printf("Error bind env: %v", err)
			return err
		}
	}
	return nil
}
