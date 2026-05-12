package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	API      APIConfig      `mapstructure:"api"`
	Panel    PanelConfig    `mapstructure:"panel"`
}

type AppConfig struct {
	Name     string `mapstructure:"name"`
	Env      string `mapstructure:"env"`
	Debug    bool   `mapstructure:"debug"`
	URL      string `mapstructure:"url"`
	Timezone string `mapstructure:"timezone"`
	Locale   string `mapstructure:"locale"`
	Secret   string `mapstructure:"secret"`
}

type DatabaseConfig struct {
	Driver    string `mapstructure:"driver"`
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Database  string `mapstructure:"database"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Charset   string `mapstructure:"charset"`
	Collation string `mapstructure:"collation"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	TwoFactorRequired bool `mapstructure:"2fa_required"`
	TwoFactorBytes    int  `mapstructure:"2fa_bytes"`
	TwoFactorWindow   int  `mapstructure:"2fa_window"`
	VerifyNewer       bool `mapstructure:"verify_newer"`
}

type APIConfig struct {
	KeyLimit      int `mapstructure:"key_limit"`
	KeyExpireTime int `mapstructure:"key_expire_time"`
}

type PanelConfig struct {
	UseBinaryPrefix            bool  `mapstructure:"use_binary_prefix"`
	EditableServerDescriptions bool  `mapstructure:"editable_server_descriptions"`
	WebhookPruneDays           int   `mapstructure:"webhook_prune_days"`
	FilesMaxEditSize           int64 `mapstructure:"files_max_edit_size"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("PANEL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "Pelican")
	v.SetDefault("app.env", "production")
	v.SetDefault("app.debug", false)
	v.SetDefault("app.url", "http://localhost")
	v.SetDefault("app.timezone", "UTC")
	v.SetDefault("app.secret", "")
	v.SetDefault("app.locale", "en")
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.database", "pelican")
	v.SetDefault("database.charset", "utf8mb4")
	v.SetDefault("database.collation", "utf8mb4_unicode_ci")
	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("auth.2fa_required", false)
	v.SetDefault("auth.2fa_bytes", 32)
	v.SetDefault("auth.2fa_window", 4)
	v.SetDefault("auth.verify_newer", true)
	v.SetDefault("api.key_limit", 25)
	v.SetDefault("api.key_expire_time", 720)
	v.SetDefault("panel.use_binary_prefix", true)
	v.SetDefault("panel.editable_server_descriptions", true)
	v.SetDefault("panel.webhook_prune_days", 30)
	v.SetDefault("panel.files_max_edit_size", 4194304)
}
