package log

import (
	"os"
	"path/filepath"

	"github.com/elastic/beats/libbeat/logp"
)

var (
	levelMap = map[string]logp.Level{
		logp.DebugLevel.String():    logp.DebugLevel,
		logp.InfoLevel.String():     logp.InfoLevel,
		logp.WarnLevel.String():     logp.WarnLevel,
		logp.ErrorLevel.String():    logp.ErrorLevel,
		logp.CriticalLevel.String(): logp.CriticalLevel,
	}

	// DefaultLogger provides global logging motheds
	DefaultLogger Logger = logp.NewLogger("nop")
)

// parseLevel parse string to logp.Level
func parseLevel(l string) logp.Level {
	if lv, exist := levelMap[l]; exist {
		return lv
	}
	return logp.InfoLevel
}

// Config initialize logp package
func Config(level, logPath string, toStderr bool, maxSize, maxBackups uint) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if logPath == "" {
		logPath = filepath.Join(wd, "logs")
	}
	l := parseLevel(level)
	config := logp.DefaultConfig()
	config.Level = l
	config.ToStderr = toStderr
	config.Files.Name = "octopus.log"
	config.Files.Path = logPath
	config.Files.MaxSize = maxSize
	config.Files.MaxBackups = maxBackups
	if err := logp.Configure(config); err != nil {
		panic(err)
	}

	DefaultLogger = logp.NewLogger("octopus")
}

// Fatal calls the same method of DefaultLogger
func Fatal(args ...interface{}) {
	DefaultLogger.Fatal(args...)
}

// Fatalf calls the same method of DefaultLogger
func Fatalf(format string, args ...interface{}) {
	DefaultLogger.Fatalf(format, args...)
}

// Panic calls the same method of DefaultLogger
func Panic(args ...interface{}) {
	DefaultLogger.Panic(args...)
}

// Panicf calls the same method of DefaultLogger
func Panicf(format string, args ...interface{}) {
	DefaultLogger.Panicf(format, args...)
}

// Debug calls the same method of DefaultLogger
func Debug(args ...interface{}) {
	DefaultLogger.Debug(args...)
}

// Debugf calls the same method of DefaultLogger
func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args...)
}

// Error calls the same method of DefaultLogger
func Error(args ...interface{}) {
	DefaultLogger.Error(args...)
}

// Errorf calls the same method of DefaultLogger
func Errorf(format string, args ...interface{}) {
	DefaultLogger.Errorf(format, args...)
}

// Info calls the same method of DefaultLogger
func Info(args ...interface{}) {
	DefaultLogger.Info(args...)
}

// Infof calls the same method of DefaultLogger
func Infof(format string, args ...interface{}) {
	DefaultLogger.Infof(format, args...)
}

// Warn calls the same method of DefaultLogger
func Warn(args ...interface{}) {
	DefaultLogger.Warn(args...)
}

// Warnf calls the same method of DefaultLogger
func Warnf(format string, args ...interface{}) {
	DefaultLogger.Warnf(format, args...)
}
