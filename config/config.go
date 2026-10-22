package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           int           `json:"port"`
	Host           string        `json:"host"`
	Proxy          string        `json:"proxy"`
	Token          string        `json:"token"`
	RequestTimeout time.Duration `json:"request_timeout"`
	CacheTTL       time.Duration `json:"cache_ttl"`
	JavDBInterval  time.Duration `json:"javdb_interval"`
}

var GlobalConfig = defaultConfig()

func defaultConfig() *Config {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			port = val
		}
	}

	proxy := os.Getenv("HTTP_PROXY")
	if proxy == "" {
		proxy = os.Getenv("HTTPS_PROXY")
	}
	if proxy == "" {
		proxy = os.Getenv("ALL_PROXY")
	}

	return &Config{
		Port:           port,
		Host:           "0.0.0.0",
		Proxy:          proxy,
		Token:          os.Getenv("TOKEN"),
		RequestTimeout: 15 * time.Second,
		CacheTTL:       24 * time.Hour,
		JavDBInterval:  3 * time.Second,
	}
}
