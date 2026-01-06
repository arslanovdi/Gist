// Package config загрузка конфигурации приложения
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

// Config структура конфигурации приложения
type Config struct {
	Project struct {
		Debug           bool          `yaml:"debug"`
		ShutdownTimeout time.Duration `yaml:"shutdownTimeout"`
		TTL             time.Duration `yaml:"ttl"` // Время хранения чатов в кэше
		AudioPath       string        `mapstructure:"audio_path"`
	} `yaml:"project"`

	Bot struct {
		Token             string        `yaml:"token"`                       // env BOT_TOKEN
		NgrokAuthToken    string        `mapstructure:"ngrok_auth_token"`    // env BOT_NGROK_AUTH_TOKEN
		NgrokDomain       string        `mapstructure:"ngrok_domain"`        // env BOT_NGROK_DOMAIN
		ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"` // env BOT_READ_HEADER_TIMEOUT
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

	LLM struct {
		Development      bool          `yaml:"development"`
		FlowTimeout      time.Duration `mapstructure:"flow_timeout"` // Тайм-аут выполнения сценария LLM
		DriftPercent     int           `mapstructure:"drift_percent"`
		SymbolPerToken   int           `mapstructure:"symbol_per_token"`
		MessagesPerBatch int           `mapstructure:"messages_per_batch"`
		DefaultProvider  string        `mapstructure:"default_provider"` // switch of Ollama, OpenRouter, Gemini, OpenAI.

		Ollama struct {
			Enabled       bool   `mapstructure:"enabled"`
			ServerAddress string `mapstructure:"server_address"`
			Timeout       int    `yaml:"timeout"`
			Model         string `yaml:"model"`
			ContextWindow int    `mapstructure:"context_window"`
		} `yaml:"Ollama"`

		OpenRouter struct {
			Enabled       bool   `mapstructure:"enabled"`
			Model         string `yaml:"model"`
			ContextWindow int    `mapstructure:"context_window"`
		} `yaml:"OpenRouter"`

		Gemini struct {
			Enabled       bool     `mapstructure:"enabled"`
			Model         string   `yaml:"model"`
			ContextWindow int      `mapstructure:"context_window"`
			ApiKeys       []string `mapstructure:"api_keys"`
		} `yaml:"Gemini"`

		OpenAI struct {
			Enabled       bool   `mapstructure:"enabled"`
			Model         string `yaml:"model"`
			ContextWindow int    `mapstructure:"context_window"`
		} `yaml:"OpenAI"`

		TTS struct {
			Gemini struct {
				Model        string `yaml:"model"`
				LanguageCode string `mapstructure:"language_code"`
				VoiceName    string `mapstructure:"voice_name"`
			} `yaml:"Gemini"`
		} `mapstructure:"tts"`
	} `yaml:"llm"`
}

// LoadConfig загружает конфигурацию приложения из YAML-файла.
//
// - Путь к конфигурационному файлу получает из переменной окружения CONFIG_FILE, если она задана.
//
// - Поддерживается переопределение любых параметров через переменные окружения (например, SERVICE_HOST заменит значение service.host из конфига)
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
