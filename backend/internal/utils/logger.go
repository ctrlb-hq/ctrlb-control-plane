package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Global Logger variable
var Logger *zap.Logger

func InitLogger() {
	// Define encoder for better readability
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "level",
		NameKey:      "logger",
		CallerKey:    "caller",
		MessageKey:   "message",
		LineEnding:   zapcore.DefaultLineEnding,
		EncodeLevel:  zapcore.CapitalColorLevelEncoder, // Colors in console logs
		EncodeTime:   zapcore.ISO8601TimeEncoder,       // Human-readable timestamps
		EncodeCaller: zapcore.ShortCallerEncoder,       // Short file:line format
	}

	// Set log level (change `DebugLevel` to `InfoLevel` in production)
	logLevel := zap.NewAtomicLevelAt(zapcore.DebugLevel)

	// Define core loggers
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig) // JSON logs for files

	// Create log file
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}

	// Define core for writing logs to both console & file
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), logLevel), // Console logging
		zapcore.NewCore(fileEncoder, zapcore.Lock(logFile), logLevel),      // File logging
	)

	// Create logger (without unnecessary stack traces)
	Logger = zap.New(core, zap.AddCaller())

	// Replace Zap's global logger
	zap.ReplaceGlobals(Logger)

	// Recover from panics and log stack trace only for panics
	defer recoverFromPanic()
}

// recoverFromPanic logs panic stack traces and exits gracefully
func recoverFromPanic() {
	if r := recover(); r != nil {
		Logger.Fatal("Panic occurred!", zap.Any("error", r), zap.Stack("stacktrace"))
	}
}
