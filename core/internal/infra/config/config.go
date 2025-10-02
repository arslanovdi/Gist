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
