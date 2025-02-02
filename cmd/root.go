package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/dimdasci/seek/internal/logging"
)

var (
	cfgFile    string
	output     string
	Version    string
	BuildTime  string
	CommitHash string
)

var rootCmd = &cobra.Command{
	Use:   "seek",
	Short: "Seek is a tool for information retrieval from the web",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logger := logging.GetLogger()
		logger.Info("Seek started",
			zap.String("version", Version),
			zap.String("build_time", BuildTime),
			zap.String("commit_hash", CommitHash),
			zap.String("config_file", viper.ConfigFileUsed()),
		)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogging)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .seek.yaml)")
	rootCmd.PersistentFlags().StringVar(&output, "output", "stdout", "output file (default is stdout)")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "logs/seek.log")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
		}
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".seek")
	}

	// Set environment variables prefix and automatically read them
	viper.SetEnvPrefix("SEEK")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Explicitly read the config file
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if we can't find a config file, but we should log other errors
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
			os.Exit(1)
		}
	}
}

func initLogging() {
	level := viper.GetString("logging.level")
	logFile := viper.GetString("logging.file")

	if err := logging.InitLogger(level, logFile); err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	logger := logging.GetLogger()
	logger.Info("Logger initialized",
		zap.String("level", level),
		zap.String("file", logFile),
	)
}
