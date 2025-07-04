package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var config Config

type Config struct {
	Postgres struct {
		Host     string `yaml:"host"`
		Port     uint   `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"database,omitempty"`
	}
	Application struct {
		Host   string `yaml:"host"`
		Prefix string `yaml:"prefix"`
		Port   string `yaml:"port"`
		Name   string `yaml:"name"`
	}
	Jwt struct {
		SecretKey string `yaml:"secret_key"`
		Issuer    string `yaml:"issuer"`
	}
	Webhook struct {
		Url string `yaml:"url"`
	}
	Usr struct {
		Count int `yaml:"count"`
	}
}

func GetConfig() *Config {
	return &config
}

func (c Config) GetPostgresDsn() string {
	return fmt.Sprintf(
		"host=%s password='%s' port=%d user=%s dbname=%s sslmode=disable TimeZone=Asia/Yekaterinburg",
		c.Postgres.Host,
		c.Postgres.Password,
		c.Postgres.Port,
		c.Postgres.User,
		c.Postgres.Database,
	)
}

func Load(configPath string) (*Config, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error read c file: %s", err)
		return nil, err
	}
	var c Config
	err = yaml.Unmarshal(content, &c)
	if err != nil {
		log.Fatalf("Error parse c file: %s", err)
		return nil, err
	}

	if c.Postgres.Database == "" {
		c.Postgres.Database = c.Postgres.User
	}
	config = c
	return &c, err
}
