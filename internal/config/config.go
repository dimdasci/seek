package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Logging struct {
		Level string
		File  string
	}
	OpenAI struct {
		APIKey     string
		Reasoning  ServiceConfig
		Completion ServiceConfig
	}
	WebSearch struct {
		Tavily struct {
			Timeout time.Duration
		}
	}
	WebRead struct {
		Timeout time.Duration
	}
}

type ServiceConfig struct {
	Model     string
	Timeout   time.Duration
	MaxTokens int64
}

var appConfig Config

func Load(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".seek")
	}

	viper.SetEnvPrefix("SEEK")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(&appConfig)
}

func setDefaults() {
	home := os.Getenv("HOME")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", filepath.Join(home, "logs", "seek.log"))

	viper.SetDefault("openai.reasoning.timeout", "60s")
	viper.SetDefault("openai.completion.timeout", "30s")
	viper.SetDefault("openai.reasoning.max_tokens", 2000)
	viper.SetDefault("openai.completion.max_tokens", 1000)

	viper.SetDefault("websearch.tavily.timeout", "10s")
	viper.SetDefault("webread.timeout", "10s")
}

func Get() *Config {
	return &appConfig
}
