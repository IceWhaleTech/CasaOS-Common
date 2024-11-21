package logger

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/samber/lo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var loggers *zap.Logger

var (
	infoDebouncer  *debouncer
	errorDebouncer *debouncer
)

// Add debouncer type
type debouncer struct {
	mu       sync.Mutex
	timer    *time.Timer
	interval time.Duration
	messages []logMessage
}

type logMessage struct {
	msg    string
	fields []zap.Field
}

func newDebouncer(interval time.Duration) *debouncer {
	return &debouncer{
		interval: interval,
	}
}

// Add these new functions
func InfoDebounced(message string, fields ...zap.Field) {
	if infoDebouncer == nil {
		infoDebouncer = newDebouncer(5 * time.Second)
	}

	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)

	infoDebouncer.mu.Lock()
	defer infoDebouncer.mu.Unlock()

	infoDebouncer.messages = append(infoDebouncer.messages, logMessage{msg: message, fields: fields})

	if infoDebouncer.timer != nil {
		infoDebouncer.timer.Stop()
	}

	infoDebouncer.timer = time.AfterFunc(infoDebouncer.interval, func() {
		infoDebouncer.mu.Lock()
		defer infoDebouncer.mu.Unlock()

		for _, msg := range infoDebouncer.messages {
			loggers.Info(msg.msg, msg.fields...)
		}
		infoDebouncer.messages = nil
	})
}

func ErrorDebounced(message string, fields ...zap.Field) {
	if errorDebouncer == nil {
		errorDebouncer = newDebouncer(5 * time.Second)
	}

	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)

	errorDebouncer.mu.Lock()
	defer errorDebouncer.mu.Unlock()

	errorDebouncer.messages = append(errorDebouncer.messages, logMessage{msg: message, fields: fields})

	if errorDebouncer.timer != nil {
		errorDebouncer.timer.Stop()
	}

	errorDebouncer.timer = time.AfterFunc(errorDebouncer.interval, func() {
		errorDebouncer.mu.Lock()
		defer errorDebouncer.mu.Unlock()

		for _, msg := range errorDebouncer.messages {
			loggers.Error(msg.msg, msg.fields...)
		}
		errorDebouncer.messages = nil
	})
}

func getFileLogWriter(logPath string, logFileName string, logFileExt string) (writeSyncer zapcore.WriteSyncer) {
	// 使用 lumberjack 实现 log rotate
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logPath, fmt.Sprintf("%s.%s", logFileName, logFileExt)),
		MaxSize:    10,
		MaxBackups: 60,
		MaxAge:     1,
		Compress:   true,
	}

	return zapcore.AddSync(lumberJackLogger)
}

func LogInitWithWriterSyncers(syncers ...zapcore.WriteSyncer) {
	encoder := getEncoder()
	loggers = zap.New(
		zapcore.NewTee(
			lo.Map(
				syncers,
				func(syncer zapcore.WriteSyncer, index int) zapcore.Core {
					return zapcore.NewCore(encoder, syncer, zapcore.InfoLevel)
				})...,
		))
}

// for unit tests
func LogInitConsoleOnly() {
	LogInitWithWriterSyncers(
		zapcore.AddSync(os.Stdout),
	)
}

func LogInit(logPath string, logFileName string, logFileExt string) {
	LogInitWithWriterSyncers(
		zapcore.AddSync(os.Stdout),
		getFileLogWriter(logPath, logFileName, logFileExt),
	)
}

func Info(message string, fields ...zap.Field) {
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	loggers.Info(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	loggers.Error(message, fields...)
}

func getCallerInfoForLog() (callerFields []zap.Field) {
	pc, file, line, ok := runtime.Caller(2) // 回溯两层，拿到写日志的调用方的函数信息
	if !ok {
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	funcName = path.Base(funcName) // Base函数返回路径的最后一个元素，只保留函数名

	callerFields = append(callerFields, zap.String("func", funcName), zap.String("file", file), zap.Int("line", line))
	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}
