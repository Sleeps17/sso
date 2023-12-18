package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env         string        `yaml:"env" env-default:"local"`
	StoragePath string        `yaml:"storage_path" env-required:"true"`
	TokenTTL    time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPCConfig  `yaml:"grpc_config"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-default:"4404""`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
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
