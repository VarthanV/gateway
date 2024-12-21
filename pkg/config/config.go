package config

import (
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Server        ServerConfig        `toml:"server"`
	Routes        []RouteConfig       `toml:"routes"`
	LoadBalancing LoadBalancingConfig `toml:"load_balancing"`
	RateLimit     RateLimitConfig     `toml:"rate_limit"`
	Logging       LoggingConfig       `toml:"logging"`
	CORS          CORSConfig          `toml:"cors"`
	JWTConfig     JWTConfig           `toml:"jwt"`
}

type ServerConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type RouteConfig struct {
	Path      string           `toml:"path"`
	Methods   []string         `toml:"methods"`
	Upstreams []UpstreamConfig `toml:"upstreams"`
}

type UpstreamConfig struct {
	URL    string `toml:"url"`
	Weight int    `toml:"weight"`
}

type LoadBalancingConfig struct {
	Algorithm string `toml:"algorithm"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `toml:"requests_per_minute"`
	BurstLimit        int `toml:"burst_limit"`
}

type LoggingConfig struct {
	Level string `toml:"level"`
	File  string `toml:"file"`
}

type CORSConfig struct {
	AllowedOrigins []string `toml:"allowed_origins"`
	AllowedMethods []string `toml:"allowed_methods"`
	AllowedHeaders []string `toml:"allowed_headers"`
}

type JWTConfig struct {
	SecretKey string `toml:"secret_key"`
}

func (c *Config) Load(filename string) {
	if _, err := toml.DecodeFile(filename, &c); err != nil {
		logrus.Fatalf("Failed to parse config: %v", err)
	}
}
