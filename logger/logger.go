package logger

import (
	"log"
	"os"
)

var (
	InfoLog  = log.New(os.Stdout, "[INFO]", log.LstdFlags|log.Lshortfile)
	ErrLog   = log.New(os.Stderr, "[ERROR]", log.LstdFlags|log.Lshortfile)
	TraceLog *log.Logger
)

func Info(format string, args ...interface{}) {
	InfoLog.Printf(format, args...)
}

func Err(format string, args ...interface{}) {
	ErrLog.Printf(format, args...)
}

func Trace(format string, args ...interface{}) {
	if TraceLog != nil {
		TraceLog.Printf(format, args...)
	}
}

func SetTraceLogger() {
	TraceLog = log.New(os.Stdout, "[TRACE]", log.LstdFlags|log.Lshortfile)
}
