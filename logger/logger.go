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
	infoLogger    *log.Logger
	successLogger *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	debugLogger   *log.Logger
)

var (
	blue   = color("\033[1;34m%s\033[0m")
	green  = color("\033[1;32m%s\033[0m")
	yellow = color("\033[1;33m%s\033[0m")
	red    = color("\033[1;31m%s\033[0m")
	purple = color("\033[1;35m%s\033[0m")
)

func Init() {
	name := fmt.Sprintf("./storage/logs/%s.txt", time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err.Error())
	}

	infoLogger = log.New(file, prefix(typeInfo), log.Ldate|log.Ltime)
	successLogger = log.New(file, prefix(typeSuccess), log.Ldate|log.Ltime)
	warningLogger = log.New(file, prefix(typeWarning), log.Ldate|log.Ltime)
	errorLogger = log.New(file, prefix(typeError), log.Ldate|log.Ltime)
	debugLogger = log.New(file, prefix(typeDebug), log.Ldate|log.Ltime)
}

func Info(v interface{}) {
	infoLogger.Println(stringify(v))
	log.Println(blue(prefix(typeInfo) + stringify(v)))
}

func Success(v interface{}) {
	successLogger.Println(stringify(v))
	log.Println(green(prefix(typeSuccess) + stringify(v)))
}

func Warning(v interface{}) {
	warningLogger.Println(stringify(v))
	log.Println(yellow(prefix(typeWarning) + stringify(v)))
}

func Error(v interface{}) {
	errorLogger.Println(stringify(v))
	log.Println(red(prefix(typeError) + stringify(v)))
}

func Debug(v interface{}) {
	debugLogger.Println(stringify(v))
	log.Println(purple(prefix(typeDebug) + stringify(v)))
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
