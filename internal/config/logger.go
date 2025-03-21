package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lj "gopkg.in/natefinch/lumberjack.v2"
)

// Load Logger configuration
func LoadLogger(logFilePath string) *zap.Logger {
	// Configure log rotation
	logWriter := zapcore.AddSync(&lj.Logger{
		Filename:   logFilePath + "/teleminio-uploader.log",
		MaxBackups: 3,
		MaxSize:    1, // megabytes
		MaxAge:     7, // days
	})

	// Set log level based on DEBUG environment variable
	logLevel := zapcore.InfoLevel
	if os.Getenv("DEBUG") == "true" {
		logLevel = zapcore.DebugLevel
	}

	// Create encoder with time format
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create core
	logCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		logWriter,
		logLevel,
	)

	// Create logger with caller info
	lg := zap.New(logCore, zap.AddCaller(), zap.AddCallerSkip(1))
	defer func() { _ = lg.Sync() }()

	return lg
}
