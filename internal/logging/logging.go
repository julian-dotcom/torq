// Package logging - Log configuration using zerolog package.
// Two configurations available: for dev environment(InitLogDev) and prod environment(InitLogProd)
package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Config Configuration for logging
type Config struct {
	// Enable console logging
	ConsoleLoggingEnabled bool
	// EncodeLogsAsJson makes the log framework log JSON
	EncodeLogsAsJson bool
	// FileLoggingEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileLoggingEnabled bool
	// Directory to log to to when filelogging is enabled
	Directory string
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int
}

type Logger struct {
	*zerolog.Logger
}

// InitLogDev - used for dev environment.
// It pretty prints log data in the console. Also, there is the option of writing the data in a file.
// In the log file data is written in json format. Choose file name and location
func InitLogDev() {
	logConfig := Config{
		ConsoleLoggingEnabled: true,
		EncodeLogsAsJson:      false,
		FileLoggingEnabled:    true,
		Directory:             "./",
		Filename:              "marketdata.log",
		MaxSize:               0,
		MaxBackups:            0,
		MaxAge:                0,
	}
	logger := Configure(logConfig)
	log.Logger = *logger.Logger
}

// InitLogProd - used for prod environment.
// It prints in the console data in json format
func InitLogProd() {
	ConfigureCallerMarshalFunc()
	log.Logger = log.With().Caller().Logger()
}

// Configure sets up the logging framework
//
// In production, the container logs will be collected and file logging should be disabled. However,
// during development it's nicer to see logs as text and optionally write to a file when debugging
// problems in the containerized pipeline
//
// The output log file will be located at /var/log/service-xyz/service-xyz.log and
// will be rolled according to configuration set.
func Configure(config Config) *Logger {
	var writers []io.Writer

	if config.ConsoleLoggingEnabled {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}
	if config.FileLoggingEnabled {
		writers = append(writers, newRollingFile(config))
	}
	mw := io.MultiWriter(writers...)

	// zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(mw).With().Timestamp().Caller().Logger()

	logger.Info().
		Bool("fileLogging", config.FileLoggingEnabled).
		Bool("jsonLogOutput", config.EncodeLogsAsJson).
		Str("logDirectory", config.Directory).
		Str("fileName", config.Filename).
		Int("maxSizeMB", config.MaxSize).
		Int("maxBackups", config.MaxBackups).
		Int("maxAgeInDays", config.MaxAge).
		Msg("logging configured")

	// uncomment to switch to shortPath in log caller
	ConfigureCallerMarshalFunc()

	return &Logger{
		Logger: &logger,
	}
}

func newRollingFile(config Config) io.Writer {
	if err := os.MkdirAll(config.Directory, 0744); err != nil {
		log.Error().Err(err).Str("path", config.Directory).Msg("can't create log directory")
		return nil
	}

	return &lumberjack.Logger{
		Filename:   path.Join(config.Directory, config.Filename),
		MaxBackups: config.MaxBackups, // files
		MaxSize:    config.MaxSize,    // megabytes
		MaxAge:     config.MaxAge,     // days
	}
}

// ConfigureCallerMarshalFunc - truncate caller path
func ConfigureCallerMarshalFunc() {
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		r, _ := regexp.Compile("[^\\\\/]+[\\\\/][^\\\\/]+$")
		shortPath := r.FindString(file)
		if shortPath != "" {
			file = shortPath
		}
		file = strings.ReplaceAll(file, "\\", "/")
		return file + ":" + strconv.Itoa(line)
	}
}
