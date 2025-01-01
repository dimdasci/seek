package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	cfgFile    string
	logger     *zap.Logger
	Version    string
	BuildTime  string
	CommitHash string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "seek",
	Short: "A CLI utility for searching the web directly from your terminal",
	Long: `A lightning-fast command-line search utility written in Go. 
Quickly search the web directly from your terminal with a clean,
POSIX-compliant interface.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// First read the config file
		if err := initConfig(); err != nil {
			return err
		}
		// Then initialize the logger
		if err := initLogger(); err != nil {
			return err
		}

		logger.Info("Seek CLI started",
			zap.String("version", Version),
			zap.String("commit", CommitHash),
			zap.String("build_time", BuildTime),
			zap.String("config_file", viper.ConfigFileUsed()),
		)
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if logger != nil {
			return logger.Sync()
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.seek.yaml)")
}

func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".seek" (without extension)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".seek")
	}

	viper.SetEnvPrefix("SEEK") // Set prefix for environment variables
	viper.AutomaticEnv()       // read in environment variables that match

	// Replace dots with underscores in env variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// If a config file is found, read it in
	err := viper.ReadInConfig()
	return err
}

func initLogger() error {
	// Set default log settings
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "./logs/seek.log")

	// Get log level from config
	level := viper.GetString("logging.level")
	logLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return err
	}

	// Get log file path from config
	logFile := viper.GetString("logging.file")

	// Ensure log directory exists
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create core
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     7, // days
			Compress:   true,
		}),
		logLevel,
	)

	// Create logger
	logger = zap.New(fileCore,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// Log the logger settings
	logger.Debug("Logger initialized",
		zap.String("level", level),
		zap.String("file", logFile),
	)

	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	return logger
}
