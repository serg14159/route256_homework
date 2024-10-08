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

// Kafka.
type Kafka struct {
	Brokers       []string `yaml:"brokers" mapstructure:"brokers"`
	ConsumerGroup string   `yaml:"consumerGroup" mapstructure:"consumerGroup"`
	Topics        []string `yaml:"topics" mapstructure:"topics"`
}

func (k *Kafka) GetBrokers() []string {
	return k.Brokers
}

func (k *Kafka) GetConsumerGroup() string {
	return k.ConsumerGroup
}

func (k *Kafka) GetTopics() []string {
	return k.Topics
}

// Config - contains all configuration parameters in config package
type Config struct {
	Project Project `yaml:"project" mapstructure:"project"`
	Kafka   Kafka   `yaml:"kafka" mapstructure:"kafka"`
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
	viper.SetDefault("project.name", "Notifier")
	viper.SetDefault("project.environment", "development")
	// Kafka
	viper.SetDefault("kafka.brokers", "localhost:9092")
	viper.SetDefault("kafka.consumerGroup", "notifier_group")
	viper.SetDefault("kafka.topics", "loms.order-events")
}

// bindEnvVariables function for bind env variables with config name.
func (c *Config) bindEnvVariables() error {
	envVars := map[string]string{
		// Project
		"project.debug":       "PROJECT_DEBUG",
		"project.name":        "PROJECT_NAME",
		"project.environment": "PROJECT_ENVIRONMENT",
		// Kafka
		"kafka.brokers":       "KAFKA_BROKERS",
		"kafka.consumerGroup": "KAFKA_CONSUMER_GROUP",
		"kafka.topics":        "KAFKA_TOPICS",
	}

	for key, env := range envVars {
		if err := viper.BindEnv(key, env); err != nil {
			log.Printf("Error bind env: %v", err)
			return err
		}
	}
	return nil
}
