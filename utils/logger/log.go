package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	ERROR LogLevel = "ERROR"
)

type LoggerStruct struct {
	Logger      *log.Logger
	LogFile     *os.File
	environment string
}

func (l *LoggerStruct) Log(message string, level LogLevel) {
	if l.environment == "prod" && level == "DEBUG" {
		return
	}

	format := fmt.Sprintf("[%s]: %s", level, message)
	l.Logger.Println(format)
}

func (l *LoggerStruct) CloseLogger() {
	l.LogFile.Close()
}

func ConfigurarLogger(filename, env string) *LoggerStruct {
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	logger := LoggerStruct{Logger: log.New(mw, "", log.LstdFlags), LogFile: logFile, environment: env}
	return &logger
}