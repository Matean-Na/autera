package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type (
	App struct {
		Env string `mapstructure:"env"`
	}

	HTTP struct {
		Addr    string        `mapstructure:"addr"`
		Timeout time.Duration `mapstructure:"timeout"`
	}

	DB struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Name     string `mapstructure:"name"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		SslMode  string `mapstructure:"ssl_mode"`
	}

	JWT struct {
		Secret string `mapstructure:"secret"`
		Issuer string `mapstructure:"issuer"`
		TTLMin int    `mapstructure:"ttl_min"`
	}

	Migrations struct {
		URL string `mapstructure:"url"`
	}
)

type Config struct {
	App        App        `mapstructure:"app"`
	HTTP       HTTP       `mapstructure:"http"`
	DB         DB         `mapstructure:"db"`
	JWT        JWT        `mapstructure:"jwt"`
	Migrations Migrations `mapstructure:"migrations"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	// дефолты
	v.SetDefault("app.env", "dev")

	v.SetDefault("http.addr", ":8080")
	v.SetDefault("http.timeout", "60s") // строкой, чтобы viper смог распарсить

	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", "5432")
	v.SetDefault("db.name", "autera")
	v.SetDefault("db.password", "123")
	v.SetDefault("db.ssl_mode", "disable")

	v.SetDefault("jwt.issuer", "autera")
	v.SetDefault("jwt.ttl_min", 120)
	v.SetDefault("jwt.secret", "")

	v.SetDefault("migrations.url", "file://migrations")

	// env: APP_ENV -> app.env и т.п.
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// normalize duration
	if cfg.HTTP.Timeout == 0 {
		// viper может не распарсить duration автоматически в time.Duration, поэтому:
		d, err := time.ParseDuration(v.GetString("http.timeout"))
		if err != nil {
			return nil, fmt.Errorf("invalid http.timeout: %w", err)
		}
		cfg.HTTP.Timeout = d
	}

	return &cfg, nil
}
