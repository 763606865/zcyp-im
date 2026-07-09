package config

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	HTTP       HTTPConfig       `mapstructure:"http"`
	WebSocket  HTTPConfig       `mapstructure:"websocket"`
	MySQL      MySQLConfig      `mapstructure:"mysql"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Moderation ModerationConfig `mapstructure:"moderation"`
}

type HTTPConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type MySQLConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Database        string `mapstructure:"database"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime_minutes"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	Issuer      string `mapstructure:"issuer"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type ModerationConfig struct {
	BlockedWords []string `mapstructure:"blocked_words"`
}

func Load() (Config, error) {
	gotenv.Load(".env", ".env.local")

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")
	v.SetEnvPrefix("ZCYP_IM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("websocket.host", "0.0.0.0")
	v.SetDefault("websocket.port", 9012)
	v.SetDefault("mysql.enabled", false)
	v.SetDefault("mysql.host", "127.0.0.1")
	v.SetDefault("mysql.port", 3306)
	v.SetDefault("mysql.database", "im")
	v.SetDefault("mysql.username", "root")
	v.SetDefault("mysql.password", "")
	v.SetDefault("mysql.charset", "utf8mb4")
	v.SetDefault("mysql.parse_time", true)
	v.SetDefault("mysql.max_open_conns", 20)
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("mysql.conn_max_lifetime_minutes", 30)
	v.SetDefault("jwt.secret", "change-me")
	v.SetDefault("jwt.issuer", "zcyp-im")
	v.SetDefault("jwt.expire_hours", 24)
	v.SetDefault("moderation.blocked_words", []string{})

	var cfg Config
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
