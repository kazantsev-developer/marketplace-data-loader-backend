// Package config управляет конфигурацией приложения.
package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
) 

// Config содержит конфигурационные параметры всего приложения.
type Config struct {
	WB       WBConfig
	DB       DBConfig
	Ozon     OzonConfig
	MS       MoyskladConfig
	Settings SettingsConfig
	Port     int    `envconfig:"PORT" default:"3000"`
	Env      string `envconfig:"NODE_ENV" default:"development"`
}

// WBConfig содержит параметры подключения к API Wildberries.
type WBConfig struct {
	Token             string        `envconfig:"API_TOKEN" required:"true"`
	BaseURL           string        `envconfig:"BASE_URL" default:"https://wildberries.ru"`
	OrdersEndpoint    string        `envconfig:"ORDERS_ENDPOINT" default:"/api/v1/supplier/orders"`
	Timeout           time.Duration `envconfig:"TIMEOUT" default:"30s"`
	MaxRetries        int           `envconfig:"MAX_RETRIES" default:"3"`
	PaginationDelayMs int           `envconfig:"PAGINATION_DELAY_MS" default:"61000"` // Ограничение частоты запросов к разделу статистики WB
	Flag              int           `envconfig:"FLAG" default:"0"`
}

// DBConfig содержит конфигурационные параметры пула соединений с БД.
type DBConfig struct {
	Host              string `envconfig:"HOST" required:"true"`
	Port              int    `envconfig:"PORT" default:"5432"`
	Name              string `envconfig:"NAME" required:"true"`
	User              string `envconfig:"USER" required:"true"`
	Password          string `envconfig:"PASSWORD" required:"true"`
	PoolMax           int    `envconfig:"POOL_MAX" default:"20"`
	PoolIdleTimeoutMs int    `envconfig:"POOL_IDLE_TIMEOUT_MS" default:"30000"`
	PoolConnTimeoutMs int    `envconfig:"POOL_CONN_TIMEOUT_MS" default:"5000"`
}

// OzonConfig содержит параметры подключения к Seller API Ozon.
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

// MoyskladConfig содержит параметры интеграции с JSON API МойСклад.
type MoyskladConfig struct {
	Token               string        `envconfig:"TOKEN" required:"true"`
	BaseURL             string        `envconfig:"BASE_URL" default:"https://moysklad.ru"`
	Timeout             time.Duration `envconfig:"TIMEOUT" default:"60s"`
	MaxRetries          int           `envconfig:"MAX_RETRIES" default:"5"`
	RetryDelayMs        int           `envconfig:"RETRY_DELAY_MS" default:"5000"`
	PaginationDelayMs   int           `envconfig:"PAGINATION_DELAY_MS" default:"2000"`
	HeavyRequestDelayMs int           `envconfig:"HEAVY_REQUEST_DELAY_MS" default:"20000"` // Ограничение лимитов на тяжелые отчеты остатков ERP
}

// SettingsConfig содержит глобальные параметры бизнес-логики синхронизации.
type SettingsConfig struct {
	BatchSize        int    `envconfig:"BATCH_SIZE" default:"1000"`
	DaysToLoad       int    `envconfig:"DAYS_TO_LOAD" default:"30"`
	ExcludeToday     bool   `envconfig:"EXCLUDE_TODAY" default:"true"`
	APILimit         int    `envconfig:"API_LIMIT" default:"80000"`
	UniqueOrderField string `envconfig:"UNIQUE_ORDER_FIELD" default:"srid"`
}

// LoadConfig выполняет инициализацию и валидацию конфигурации из переменных окружения.
func LoadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки конфигурации: %w", err)
	}

	if cfg.WB.Token == "" {
		return nil, fmt.Errorf("WB_API_TOKEN не может быть пустым")
	}
	if cfg.Ozon.ClientID == "" || cfg.Ozon.APIKey == "" {
		return nil, fmt.Errorf("OZON_CLIENT_ID и OZON_API_KEY обязательны")
	}
	if cfg.MS.Token == "" {
		return nil, fmt.Errorf("MS_TOKEN обязателен")
	}

	if cfg.DB.Port < 1 || cfg.DB.Port > 65535 {
		return nil, fmt.Errorf("DB_PORT должен быть в диапазоне 1-65535, получен %d", cfg.DB.Port)
	}

	return &cfg, nil
}
