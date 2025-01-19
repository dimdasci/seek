package cmd

import (
	"os"
	"path/filepath"

	"github.com/dimdasci/seek/internal/config"
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
	Long: `A command-line search utility written in Go. 
Quickly search the web directly from your terminal with a clean,
POSIX-compliant interface.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Load(cfgFile); err != nil {
			return err
		}
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

// init cobra global flags
func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.seek.yaml)")
}

// initLogger initializes the logger
func initLogger() error {
	cfg := config.Get()
	logLevel, err := zapcore.ParseLevel(cfg.Logging.Level)
	if err != nil {
		return err
	}

	logFile := cfg.Logging.File
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
		zap.String("level", cfg.Logging.Level),
		zap.String("file", cfg.Logging.File),
	)

	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	return logger
}
