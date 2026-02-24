package log

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/evilCYH/NodeHub/internal/utils"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Name    string `json:"-"`
}

var (
	wsChannel chan LogEntry

	basePath string = "build"

	useConsole bool

	useFile bool

	logger *Logger

	separator = []byte(",")
)
var consoleEncoder = zapcore.EncoderConfig{
	TimeKey:       "time",
	LevelKey:      "level",
	MessageKey:    "msg",
	CallerKey:     "caller",
	StacktraceKey: "stacktrace",
	EncodeLevel:   zapcore.CapitalColorLevelEncoder,
	EncodeTime:    zapcore.RFC3339TimeEncoder,
	EncodeCaller:  zapcore.ShortCallerEncoder,
}
var fileEncoder = zapcore.EncoderConfig{
	TimeKey:     "time",
	LevelKey:    "level",
	MessageKey:  "msg",
	EncodeLevel: zapcore.LowercaseLevelEncoder,
	EncodeTime:  zapcore.RFC3339TimeEncoder,
}

type Logger struct {
	*zap.SugaredLogger
	bufferedWriter *zapcore.BufferedWriteSyncer
}
type Config struct {
	Level      string
	Path       string
	UseConsole bool
	UseFile    bool
	Name       string
	CallerSkip int
}

func webSocketHook(entry zapcore.Entry) error {
	if wsChannel == nil {
		return nil
	}

	logEntry := LogEntry{
		Level:   entry.Level.String(),
		Message: entry.Message,
		Name:    entry.LoggerName,
	}

	select {
	case wsChannel <- logEntry:
	default:
	}

	return nil
}

func init() {
	wsChannel = make(chan LogEntry, 1000)

	logger, _ = NewLogger(Config{
		Level:      "debug",
		UseConsole: true,
		CallerSkip: 1,
		UseFile:    false,
		Name:       "main",
	})
}

func Initialize(level, path, method string) error {
	logger.Close()

	basePath = path

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		if err := os.MkdirAll(basePath, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	mainPath := filepath.Join(basePath, "main", time.Now().Format("20060102150405")+".log")

	switch method {
	case "console":
		useConsole = true
		useFile = false
	case "file":
		useConsole = false
		useFile = true
	case "both":
		useConsole = true
		useFile = true
	default:
		useConsole = true
		useFile = false
	}

	var err error
	logger, err = NewLogger(Config{
		Level:      level,
		Path:       mainPath,
		UseConsole: useConsole,
		UseFile:    useFile,
		Name:       "main",
		CallerSkip: 1,
	})
	if err != nil {
		return err
	}
	return nil
}
func GetDefaultLogger() *Logger {
	return logger
}
func NewTaskLogger(name string, taskid uint16, level string, writeFile bool) (*Logger, error) {
	taskidstr := strconv.FormatUint(uint64(taskid), 10)
	loggerName := "task_" + name + "_" + taskidstr
	path := filepath.Join(basePath, name, taskidstr, time.Now().Format("20060102150405")+".log")
	return NewLogger(Config{
		Level:      level,
		Path:       path,
		UseConsole: utils.IsDebug(),
		UseFile:    writeFile,
		Name:       loggerName,
		CallerSkip: 1,
	})
}

func GetWSChannel() <-chan LogEntry {
	return wsChannel
}

func NewLogger(config Config) (*Logger, error) {
	parsedLevel, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		parsedLevel = zapcore.InfoLevel
	}

	var cores []zapcore.Core
	var bufferedWriter *zapcore.BufferedWriteSyncer

	if config.UseConsole {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoder),
			zapcore.AddSync(os.Stdout),
			parsedLevel,
		)
		cores = append(cores, consoleCore)
	}

	if config.UseFile && config.Path != "" {
		file, err := createLogFile(config.Path)
		if err != nil {
			return nil, err
		}
		bufferedWriter = &zapcore.BufferedWriteSyncer{
			WS: zapcore.AddSync(file),
		}
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(fileEncoder),
			bufferedWriter,
			parsedLevel,
		)
		cores = append(cores, fileCore)
	}

	wsEncoderConfig := zapcore.EncoderConfig{
		LevelKey:    "level",
		MessageKey:  "msg",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	}

	wsCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(wsEncoderConfig),
		zapcore.AddSync(io.Discard),
		zapcore.DebugLevel,
	)
	cores = append(cores, wsCore)

	core := zapcore.NewTee(cores...)
	logger := zap.New(
		core,
		zap.Hooks(webSocketHook),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCallerSkip(config.CallerSkip),
		zap.AddCaller(),
	)
	logger.Named(config.Name)

	return &Logger{
		SugaredLogger:  logger.Sugar(),
		bufferedWriter: bufferedWriter,
	}, nil
}

func createLogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return file, nil
}

func (l *Logger) Close() error {
	l.SugaredLogger.Sync()

	if l.bufferedWriter != nil {
		if err := l.bufferedWriter.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to flush buffered writer: %v\n", err)
		}
		l.bufferedWriter = nil
	}

	return nil
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}
func Info(args ...interface{}) {
	logger.Info(args...)
}
func Warn(args ...interface{}) {
	logger.Warn(args...)
}
func Error(args ...interface{}) {
	logger.Error(args...)
}
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}
func Close() error {
	return logger.Close()
}

func CleanupOldLogs(retentionDays int) error {
	if retentionDays <= 0 {
		return fmt.Errorf("retention days must be greater than 0")
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	return filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			Warnf("Error accessing path %s: %v", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".log") {
			return nil
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(path); err != nil {
				Warnf("Failed to remove old log file %s: %v", path, err)
			} else {
				Infof("Removed old log file: %s", path)
			}
		}

		return nil
	})
}

func GetLogFileList(path string) ([]uint64, error) {
	var timestamps []uint64

	fullPath := filepath.Join(basePath, path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return timestamps, nil
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".log") {
			continue
		}

		timeStr := strings.TrimSuffix(filename, ".log")

		if len(timeStr) != 14 {
			continue
		}

		timestamp, err := strconv.ParseUint(timeStr, 10, 64)
		if err != nil {
			continue
		}

		if _, err := time.Parse("20060102150405", timeStr); err != nil {
			continue
		}

		timestamps = append(timestamps, timestamp)
	}

	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i] > timestamps[j]
	})
	return timestamps, nil
}
func StreamLogToHTTP(path string, timestamp uint64, writer io.Writer) error {
	timeStr := strconv.FormatUint(timestamp, 10)
	if len(timeStr) != 14 {
		return fmt.Errorf("invalid timestamp format: %d", timestamp)
	}

	if _, err := time.Parse("20060102150405", timeStr); err != nil {
		return fmt.Errorf("invalid timestamp: %d", timestamp)
	}

	filename := timeStr + ".log"
	fullPath := filepath.Join(basePath, path, filename)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("log file not found: %s", filename)
		}
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	flusher, canFlush := writer.(http.Flusher)

	isFirstLine := true

	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		if len(lineBytes) == 0 {
			continue
		}

		if !isFirstLine {
			if _, err := writer.Write(separator); err != nil {
				return fmt.Errorf("failed to write separator: %w", err)
			}
		}

		if _, err := writer.Write(lineBytes); err != nil {
			return fmt.Errorf("failed to write log line: %w", err)
		}

		if canFlush {
			flusher.Flush()
		}

		isFirstLine = false
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}
func DeleteLog(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	fullPath := filepath.Join(basePath, path)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}

	Debugf("Successfully removed log dir: %s", path)
	return nil
}
