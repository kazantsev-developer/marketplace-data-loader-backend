// Package config loads and validates application configuration from environment variables
package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration parameters for the application
type Config struct {
	WB       WBConfig       `envconfig:"WB"`
	DB       DBConfig       `envconfig:"DB"`
	Ozon     OzonConfig     `envconfig:"OZON"`
	MS       MoyskladConfig `envconfig:"MS"`
	Settings SettingsConfig `envconfig:"SETTINGS"`
	Port     int            `envconfig:"PORT" default:"3000"`
	Env      string         `envconfig:"APP_ENV" default:"development"`
}

// WBConfig contains Wildberries API connection parameters
type WBConfig struct {
	Token             string        `envconfig:"API_TOKEN" required:"true"`
	BaseURL           string        `envconfig:"BASE_URL" default:"https://wildberries.ru"`
	OrdersEndpoint    string        `envconfig:"ORDERS_ENDPOINT" default:"/api/v1/supplier/orders"`
	Timeout           time.Duration `envconfig:"TIMEOUT" default:"30s"`
	MaxRetries        int           `envconfig:"MAX_RETRIES" default:"3"`
	PaginationDelayMs int           `envconfig:"PAGINATION_DELAY_MS" default:"61000"`
	Flag              int           `envconfig:"FLAG" default:"0"`
}

// DBConfig contains PostgreSQL connection pool parameters
type DBConfig struct {
	Host              string `envconfig:"HOST" required:"true"`
	Port              int    `envconfig:"PORT" default:"5432"`
	Name              string `envconfig:"NAME" required:"true"`
	User              string `envconfig:"USER" required:"true"`
	Password          string `envconfig:"PASSWORD" required:"true"`
	SSLMode           string `envconfig:"SSLMODE" default:"disable"`
	PoolMax           int    `envconfig:"POOL_MAX" default:"20"`
	PoolIdleTimeoutMs int    `envconfig:"POOL_IDLE_TIMEOUT_MS" default:"30000"`
	PoolConnTimeoutMs int    `envconfig:"POOL_CONN_TIMEOUT_MS" default:"5000"`
}

// OzonConfig contains Ozon Seller API connection parameters
type OzonConfig struct {
	ClientID          string        `envconfig:"CLIENT_ID" required:"true"`
	APIKey            string        `envconfig:"API_KEY" required:"true"`
	BaseURL           string        `envconfig:"BASE_URL" default:"https://ozon.ru"`
	FBOEndpoint       string        `envconfig:"FBO_ENDPOINT" default:"/v2/posting/fbo/list"`
	FBSEndpoint       string        `envconfig:"FBS_ENDPOINT" default:"/v3/posting/fbs/list"`
	Limit             int           `envconfig:"LIMIT" default:"1000"`
	Timeout           time.Duration `envconfig:"TIMEOUT" default:"30s"`
	PaginationDelayMs int           `envconfig:"PAGINATION_DELAY_MS" default:"200"`
}

// MoyskladConfig contains MoySklad JSON API integration parameters
type MoyskladConfig struct {
	Token               string        `envconfig:"TOKEN" required:"true"`
	BaseURL             string        `envconfig:"BASE_URL" default:"https://moysklad.ru"`
	Timeout             time.Duration `envconfig:"TIMEOUT" default:"60s"`
	MaxRetries          int           `envconfig:"MAX_RETRIES" default:"5"`
	RetryDelayMs        int           `envconfig:"RETRY_DELAY_MS" default:"5000"`
	PaginationDelayMs   int           `envconfig:"PAGINATION_DELAY_MS" default:"2000"`
	HeavyRequestDelayMs int           `envconfig:"HEAVY_REQUEST_DELAY_MS" default:"20000"`
}

// SettingsConfig holds global synchronization parameters
type SettingsConfig struct {
	BatchSize        int    `envconfig:"BATCH_SIZE" default:"1000"`
	DaysToLoad       int    `envconfig:"DAYS_TO_LOAD" default:"30"`
	ExcludeToday     bool   `envconfig:"EXCLUDE_TODAY" default:"true"`
	APILimit         int    `envconfig:"API_LIMIT" default:"80000"`
	UniqueOrderField string `envconfig:"UNIQUE_ORDER_FIELD" default:"srid"`
}

// LoadConfig reads configuration from environment variables with proper prefixes
func LoadConfig() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("process env config: %w", err)
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid app port: %d", cfg.Port)
	}
	if cfg.DB.Port < 1 || cfg.DB.Port > 65535 {
		return nil, fmt.Errorf("invalid db port: %d", cfg.DB.Port)
	}

	return &cfg, nil
}
