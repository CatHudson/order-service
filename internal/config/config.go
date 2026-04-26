package config

import (
	"fmt"
	"net"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GRPC     GRPCConfig        `yaml:"grpc"`
	Postgres PrimaryAndReplica `yaml:"postgres"`
	Kafka    KafkaConfig       `yaml:"kafka"`
	Redis    RedisConfig       `yaml:"redis"`
	Asynq    AsynqConfig       `yaml:"asynq"`
}

type AsynqConfig struct {
	Concurrency int `yaml:"concurrency"`
}

type RedisConfig struct {
	Addr string `yaml:"address"`
}

type KafkaConfig struct {
	Address   string              `yaml:"address"`
	Producers KafkaProducerConfig `yaml:"producers"`
	Consumers KafkaConsumerConfig `yaml:"consumers"`
}

type KafkaConsumerConfig struct {
	CreateOrderTopic string `yaml:"createOrderTopic"`
	OrderResultTopic string `yaml:"orderResultTopic"`
	GroupID          string `yaml:"groupId"`
}

type KafkaProducerConfig struct {
	CreateOrderTopic string `yaml:"createOrderTopic"`
	OrderResultTopic string `yaml:"orderResultTopic"`
	DLQ              string `yaml:"dlq"`
}

type GRPCConfig struct {
	Network string `yaml:"network"`
	Port    int    `yaml:"port"`
}

type PrimaryAndReplica struct {
	Primary PostgresConfig `yaml:"primary"`
	Replica PostgresConfig `yaml:"replica"`
}

type PostgresConfig struct {
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

func (p PostgresConfig) DSN() string {
	hostPort := net.JoinHostPort(p.Host, fmt.Sprintf("%d", p.Port))
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		p.User, p.Password, hostPort, p.Database, p.SSLMode,
	)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not load config: %w", err)
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("could not parse config: %w", err)
	}

	return &config, nil
}
