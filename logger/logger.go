package logger

import (
	"fmt"
	"github.com/memochou1993/github-rankings/util"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

var (
	InfoLogger    *log.Logger
	SuccessLogger *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
	DebugLogger   *log.Logger
)

const (
	typeInfo    = "INFO"
	typeSuccess = "SUCCESS"
	typeWarning = "WARNING"
	typeError   = "ERROR"
	typeDebug   = "DEBUG"
)

var (
	Blue   = color("\033[1;34m%s\033[0m")
	Green  = color("\033[1;32m%s\033[0m")
	Yellow = color("\033[1;33m%s\033[0m")
	Red    = color("\033[1;31m%s\033[0m")
	Purple = color("\033[1;35m%s\033[0m")
)

func Init() {
	name := fmt.Sprintf("./storage/logs/%s.txt", time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, prefix(typeInfo), log.Ldate|log.Ltime)
	SuccessLogger = log.New(file, prefix(typeSuccess), log.Ldate|log.Ltime)
	WarningLogger = log.New(file, prefix(typeWarning), log.Ldate|log.Ltime)
	ErrorLogger = log.New(file, prefix(typeError), log.Ldate|log.Ltime)
	DebugLogger = log.New(file, prefix(typeDebug), log.Ldate|log.Ltime)
}

func Info(v interface{}) {
	InfoLogger.Println(stringify(v))
	log.Println(Blue(prefix(typeInfo) + stringify(v)))
}

func Success(v interface{}) {
	SuccessLogger.Println(stringify(v))
	log.Println(Green(prefix(typeSuccess) + stringify(v)))
}

func Warning(v interface{}) {
	WarningLogger.Println(stringify(v))
	log.Println(Yellow(prefix(typeWarning) + stringify(v)))
}

func Error(v interface{}) {
	ErrorLogger.Println(stringify(v))
	log.Println(Red(prefix(typeError) + stringify(v)))
}

func Debug(v interface{}) {
	DebugLogger.Println(stringify(v))
	log.Println(Purple(prefix(typeDebug) + stringify(v)))
}

func stringify(v interface{}) string {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Struct:
		return fmt.Sprintf("%s: \"%s\"", reflect.TypeOf(v).Name(), util.JoinStruct(v, ", "))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func prefix(prefix string) string {
	return fmt.Sprintf("[%s.%s] ", strings.ToUpper(os.Getenv("APP_ENV")), prefix)
}

func color(color string) func(...interface{}) string {
	return func(args ...interface{}) string {
		return fmt.Sprintf(color, fmt.Sprint(args...))
	}
}
