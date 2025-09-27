package config

import (
	"fmt"
	"os"
	"strconv"

	"log"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Service struct {
		Name string `yaml:"name" env:"SERVICE_NAME"`
		Host string `yaml:"host" env:"SERVICE_HOST"`
		Port int32  `yaml:"port" env:"SERVICE_PORT"`
		Env  Env    `yaml:"env"  env:"SERVICE_ENV"`
	} `yaml:"service"`
	DB struct {
		Host     string `yaml:"host" env:"DB_HOST"`
		Name     string `yaml:"name" env:"DB_NAME"`
		Port     int32  `yaml:"port" env:"DB_PORT"`
		User     string `yaml:"user" env:"DB_USER"`
		Password string `yaml:"password" env:"DB_PASSWORD"`
	} `yaml:"db"`
}

type Env string

const (
	Local       Env = "local"
	Development Env = "development"
	Production  Env = "production"
)

func LoadConfig() (*Config, error) {
	config, err := getConfigFromYaml()
	if err != nil {
		log.Print(err)
	} else {
		return config, nil
	}

	config, err = getConfigFromEnv()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func getConfigFromYaml() (*Config, error) {
	var config Config

	file, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read yaml config file: %w", err)
	}

	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml config file: %w", err)
	}

	return &config, nil
}

func getConfigFromEnv() (*Config, error) {
	getEnv := func(key string, required bool) (string, error) {
		val := os.Getenv(key)
		if required && val == "" {
			return "", fmt.Errorf("%s is empty", key)
		}
		return val, nil
	}

	cfg := &Config{}

	var err error
	if cfg.Service.Name, err = getEnv("SERVICE_NAME", true); err != nil {
		return nil, err
	}
	if cfg.Service.Host, err = getEnv("SERVICE_HOST", true); err != nil {
		return nil, err
	}
	if portStr, err := getEnv("SERVICE_PORT", true); err != nil {
		return nil, err
	} else {
		port, convErr := strconv.Atoi(portStr)
		if convErr != nil {
			return nil, fmt.Errorf("invalid SERVICE_PORT: %w", convErr)
		}
		cfg.Service.Port = int32(port)
	}
	if envStr, err := getEnv("SERVICE_ENV", true); err != nil {
		return nil, err
	} else {
		cfg.Service.Env = Env(envStr)
	}

	if cfg.DB.Name, err = getEnv("DB_NAME", true); err != nil {
		return nil, err
	}
	if cfg.DB.Host, err = getEnv("DB_HOST", true); err != nil {
		return nil, err
	}
	if portStr, err := getEnv("DB_PORT", true); err != nil {
		return nil, err
	} else {
		port, convErr := strconv.Atoi(portStr)
		if convErr != nil {
			return nil, fmt.Errorf("invalid DB_PORT: %w", convErr)
		}
		cfg.DB.Port = int32(port)
	}
	if cfg.DB.User, err = getEnv("DB_USER", true); err != nil {
		return nil, err
	}
	if cfg.DB.Password, err = getEnv("DB_PASSWORD", true); err != nil {
		return nil, err
	}

	return cfg, nil
}
