package logger

import (
	"fmt"
	"github.com/memochou1993/gh-rankings/util"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	typeInfo    = "INFO"
	typeSuccess = "SUCCESS"
	typeWarning = "WARNING"
	typeError   = "ERROR"
	typeDebug   = "DEBUG"
)

var (
	blue   = color("\033[1;34m%s\033[0m")
	green  = color("\033[1;32m%s\033[0m")
	yellow = color("\033[1;33m%s\033[0m")
	red    = color("\033[1;31m%s\033[0m")
	purple = color("\033[1;35m%s\033[0m")
)

var (
	logger *Logger
)

type Logger struct {
	timestamp string
	output    *os.File
	info      *log.Logger
	success   *log.Logger
	warning   *log.Logger
	error     *log.Logger
	debug     *log.Logger
}

func (l *Logger) update() {
	if l.timestamp == now() {
		return
	}
	newLogger := newLogger()
	logger.timestamp = newLogger.timestamp
	logger.output = newLogger.output
	logger.info.SetOutput(logger.output)
	logger.success.SetOutput(logger.output)
	logger.warning.SetOutput(logger.output)
	logger.error.SetOutput(logger.output)
	logger.debug.SetOutput(logger.output)
}

func init() {
	logger = newLogger()
}

func Info(v interface{}) {
	logger.update()
	logger.info.Println(stringify(v))
	log.Println(blue(prefix(typeInfo) + stringify(v)))
}

func Success(v interface{}) {
	logger.update()
	logger.success.Println(stringify(v))
	log.Println(green(prefix(typeSuccess) + stringify(v)))
}

func Warning(v interface{}) {
	logger.update()
	logger.warning.Println(stringify(v))
	log.Println(yellow(prefix(typeWarning) + stringify(v)))
}

func Error(v interface{}) {
	logger.update()
	logger.error.Println(stringify(v))
	log.Println(red(prefix(typeError) + stringify(v)))
}

func Debug(v interface{}) {
	logger.update()
	logger.debug.Println(stringify(v))
	log.Println(purple(prefix(typeDebug) + stringify(v)))
}

func newLogger() *Logger {
	timestamp := now()
	name := fmt.Sprintf("%s/storage/logs/%s.txt", util.Root(), timestamp)
	output, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err.Error())
	}
	flag := log.Ldate | log.Ltime
	return &Logger{
		timestamp: timestamp,
		output:    output,
		info:      log.New(output, prefix(typeInfo), flag),
		success:   log.New(output, prefix(typeSuccess), flag),
		warning:   log.New(output, prefix(typeWarning), flag),
		error:     log.New(output, prefix(typeError), flag),
		debug:     log.New(output, prefix(typeDebug), flag),
	}
}

func now() string {
	return time.Now().Format("2006-01-02_15")
}

func stringify(v interface{}) string {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Struct:
		return fmt.Sprintf("%s: %s", reflect.TypeOf(v).Name(), strconv.Quote(util.ParseStruct(v, ", ")))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func prefix(prefix string) string {
	return fmt.Sprintf("[%s.%s] ", strings.ToUpper(os.Getenv("APP_ENV")), prefix)
}

func color(color string) func(...interface{}) string {
	return func(a ...interface{}) string {
		return fmt.Sprintf(color, fmt.Sprint(a...))
	}
}
