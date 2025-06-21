package logger

import (
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

// InitLogger инициализирует глобальный логгер
func InitLogger(production bool) {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Sync очищает буферы логгера
func Sync() {
	// Для простого logger'а ничего не нужно делать
}

// Exportируемые функции логгирования

func Info(msg string) {
	if InfoLogger != nil {
		InfoLogger.Println(msg)
	}
}

func Infof(format string, args ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf(format, args...)
	}
}

func Error(msg string) {
	if ErrorLogger != nil {
		ErrorLogger.Println(msg)
	}
}

func Errorf(format string, args ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf(format, args...)
	}
}

func Warn(msg string) {
	if InfoLogger != nil {
		InfoLogger.Println("WARN: " + msg)
	}
}

func Warnf(format string, args ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf("WARN: "+format, args...)
	}
}

func Fatal(msg string) {
	if ErrorLogger != nil {
		ErrorLogger.Println("FATAL: " + msg)
	}
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf("FATAL: "+format, args...)
	}
	os.Exit(1)
}
