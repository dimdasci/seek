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
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"logging"`
	OpenAI struct {
		APIKey     string        `yaml:"api_key"`
		Reasoning  ServiceConfig `yaml:"reasoning"`
		Completion ServiceConfig `yaml:"completion"`
	} `yaml:"openai"`
	WebSearch struct {
		Tavily struct {
			Timeout    time.Duration `yaml:"timeout"`
			APIKey     string        `yaml:"api_key"`
			SearchURL  string        `yaml:"search_url"`
			ExtractURL string        `yaml:"extract_url"`
			MaxResults int           `yaml:"max_results"`
		} `yaml:"tavily"`
	} `yaml:"websearch"`
	WebRead struct {
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"webreader"`
}

type ServiceConfig struct {
	Model     string        `yaml:"model"`
	Timeout   time.Duration `yaml:"timeout"`
	MaxTokens int64         `yaml:"max_tokens"`
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
