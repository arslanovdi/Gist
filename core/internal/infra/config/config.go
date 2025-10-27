package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const defaultConfigFilePath = "configs/values_local.yaml"

type Config struct {
	Project struct {
		Debug           bool          `yaml:"debug"`
		ShutdownTimeout time.Duration `yaml:"shutdownTimeout"`
	} `yaml:"project"`

	Bot struct {
		Token          string `yaml:"token"`                    // env BOT_TOKEN
		NgrokAuthToken string `mapstructure:"ngrok_auth_token"` // env BOT_NGROK_AUTH_TOKEN
		NgrokDomain    string `mapstructure:"ngrok_domain"`     // env BOT_NGROK_DOMAIN
	} `yaml:"bot"`

	Client struct {
		AppID          int           `mapstructure:"app_id"`   // env CLIENT_APP_ID	// mapstructure вместо yaml, viper некорректно парсит yaml тэги со знаком "_"
		AppHash        string        `mapstructure:"app_hash"` // env CLIENT_APP_HASH
		UserID         int64         `mapstructure:"user_id"`  // env CLIENT_USER_ID
		Phone          string        `yaml:"phone"`            // env CLIENT_PHONE
		SessionTTL     time.Duration `yaml:"sessionTTL"`
		RequestTimeout time.Duration `yaml:"requestTimeout"`
	} `yaml:"client"`

	Settings struct {
		ChatUnreadThreshold int `mapstructure:"chat_unread_threshold"`
	} `yaml:"settings"`
}

func LoadConfig() (*Config, error) {

	v := viper.New()

	v.SetConfigType("yaml")

	// автоматическая замена параметров из конфигурации соответствующей переменной окружения, при наличии, например SERVICE_HOST.
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Получаем путь до конфигурации
	configFile, ok := os.LookupEnv("CONFIG_FILE")
	if !ok {
		configFile = defaultConfigFilePath
	}
	v.SetConfigName(filepath.Base(configFile))
	v.AddConfigPath(filepath.Dir(configFile))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("fatal error config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &cfg, nil
}
