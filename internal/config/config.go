package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env      string         `yaml:"env" env-default:"prod"`
	TokenTTL time.Duration  `yaml:"token_ttl" env-required:"true"`
	Server   GrpcConfig     `yaml:"server" env-required:"true"`
	Storage  PostgresConfig `yaml:"storage"`
}

type GrpcConfig struct {
	Port    int           `yaml:"port" env-default:"4404"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type SqliteConfig struct {
	StoragePath string `yaml:"storage_path" env-required:"true"`
}

type PostgresConfig struct {
	Host     string        `yaml:"host" env-default:"localhost"`
	Port     int           `yaml:"port" env-default:"5432"`
	User     string        `yaml:"user" env-required:"true"`
	Password string        `yaml:"password" env-required:"true"`
	Database string        `yaml:"database" env-required:"true"`
	Timeout  time.Duration `yaml:"timeout"`
}

func (c *PostgresConfig) Url() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Database,
	)
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		panic("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); err != nil {
		panic("config file does not exists:" + configPath)
	}

	return MustLoadByPath(configPath)
}

func MustLoadByPath(configPath string) *Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); os.IsNotExist(err) {
		panic("can't read config: " + err.Error())
	}

	return &cfg
}
