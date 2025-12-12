package logging

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type LogLevel uint8

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
)

type Logger struct {
	level   LogLevel
	output  *os.File
	logChan chan string
	wg      sync.WaitGroup
}

func NewLogger(level LogLevel, path string) (*Logger, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	l := &Logger{
		level:   level,
		output:  file,
		logChan: make(chan string, 10000), // Large buffer for bursts
	}

	l.wg.Add(1)
	go l.worker()

	return l, nil
}

func (l *Logger) worker() {
	defer l.wg.Done()
	for line := range l.logChan {
		l.output.WriteString(line)
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := l.levelString(level)
	message := fmt.Sprintf(format, args...)

	line := fmt.Sprintf("[%s] [%s] %s\n", timestamp, levelStr, message)

	select {
	case l.logChan <- line:
	default:
		// Drop log if buffer full to avoid blocking hot path
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

func (l *Logger) Critical(format string, args ...interface{}) {
	l.log(LevelCritical, format, args...)
}

func (l *Logger) levelString(level LogLevel) string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

func (l *Logger) Close() error {
	close(l.logChan)
	l.wg.Wait()
	return l.output.Close()
}

var GlobalLogger *Logger

func InitGlobalLogger(level LogLevel, path string) error {
	logger, err := NewLogger(level, path)
	if err != nil {
		return err
	}
	GlobalLogger = logger
	return nil
}

func Debug(format string, args ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Debug(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Info(format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Warn(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Error(format, args...)
	}
}

func Critical(format string, args ...interface{}) {
	if GlobalLogger != nil {
		GlobalLogger.Critical(format, args...)
	}
}
