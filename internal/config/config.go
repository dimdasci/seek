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
		Google struct {
			Timeout    time.Duration `yaml:"timeout"`
			APIKey     string        `yaml:"api_key"`
			CX         string        `yaml:"cx"`
			SearchURL  string        `yaml:"search_url"`
			MaxResults int           `yaml:"max_results"`
		} `yaml:"google"`
	} `yaml:"websearch"`
	WebReader struct {
		Timeout          time.Duration `yaml:"timeout"`
		MinContentLength int           `yaml:"min_content_length"`
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

	setValues()

	return nil
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
	viper.SetDefault("webreader.timeout", "10s")
	viper.SetDefault("webreader.min_content_length", 128)

	viper.SetDefault("websearch.google.timeout", "5s")
	viper.SetDefault("websearch.google.max_results", 10)
}

func setValues() {
	appConfig.Logging.Level = viper.GetString("logging.level")
	appConfig.Logging.File = viper.GetString("logging.file")

	appConfig.OpenAI.APIKey = viper.GetString("openai.api_key")
	appConfig.OpenAI.Reasoning.Model = viper.GetString("openai.reasoning.model")
	appConfig.OpenAI.Reasoning.Timeout = viper.GetDuration("openai.reasoning.timeout")
	appConfig.OpenAI.Reasoning.MaxTokens = viper.GetInt64("openai.reasoning.max_tokens")
	appConfig.OpenAI.Completion.Model = viper.GetString("openai.completion.model")
	appConfig.OpenAI.Completion.Timeout = viper.GetDuration("openai.completion.timeout")
	appConfig.OpenAI.Completion.MaxTokens = viper.GetInt64("openai.completion.max_tokens")

	appConfig.WebSearch.Tavily.Timeout = viper.GetDuration("websearch.tavily.timeout")
	appConfig.WebSearch.Tavily.APIKey = viper.GetString("websearch.tavily.api_key")
	appConfig.WebSearch.Tavily.SearchURL = viper.GetString("websearch.tavily.search_url")
	appConfig.WebSearch.Tavily.ExtractURL = viper.GetString("websearch.tavily.extract_url")
	appConfig.WebSearch.Tavily.MaxResults = viper.GetInt("websearch.tavily.max_results")

	appConfig.WebSearch.Google.Timeout = viper.GetDuration("websearch.google.timeout")
	appConfig.WebSearch.Google.APIKey = viper.GetString("websearch.google.api_key")
	appConfig.WebSearch.Google.CX = viper.GetString("websearch.google.cx")
	appConfig.WebSearch.Google.SearchURL = viper.GetString("websearch.google.search_url")
	appConfig.WebSearch.Google.MaxResults = viper.GetInt("websearch.google.max_results")

	appConfig.WebReader.Timeout = viper.GetDuration("webreader.timeout")
	appConfig.WebReader.MinContentLength = viper.GetInt("webreader.min_content_length")
}

func Get() *Config {
	return &appConfig
}
