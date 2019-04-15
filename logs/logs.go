package logs

/* 日志，默认输出到consle, 调用SetLogFile 指定日志输出文件，可以按天自动生成新的文件
注意： 调用SetLogFile，必须对应的调用CloseFile
*/
import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"time"
)

type Level uint8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

const DefaultLevel Level = InfoLevel

type Logs struct {
	*log.Logger
	level    Level
	filename string //log文件名称
	f        *os.File
}

func (l *Logs) SetLevel(level Level) {
	l.level = level
}

func (l *Logs) ParseLevel(level string) {
	switch level {
	case "debug":
		l.level = DebugLevel
	case "info":
		l.level = InfoLevel
	case "warn":
		l.level = WarnLevel
	case "error":
		l.level = ErrorLevel
	case "fatal":
		l.level = FatalLevel
	case "panic":
		l.level = PanicLevel
	default:
		l.level = InfoLevel
	}
}

func (l *Logs) GetLevel() string {
	switch l.level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	}
	return ""
}

//在文件模式下检查文件，目的是为了可以按天生成log文件
func (l *Logs) checkFile() {
	if len(l.filename) <= 0 {
		return
	}

	fname := fmt.Sprintf("logs/%s%s.log", l.filename, time.Now().Format("20060102"))
	_, err := os.Stat(fname)

	//文件已经存在
	if err == nil || os.IsExist(err) {
		//文件已经打开
		if l.f != nil {
			return
		}
	}

	if l.f != nil {
		l.f.Close()
		l.f = nil
	}

	file, err := os.OpenFile(fname, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	l.f = file
	l.SetOutput(l.f)

}

func (l *Logs) SetLogFile(name string) {
	l.filename = name
}

func (l *Logs) Debug(format string, v ...interface{}) {
	if l.level > DebugLevel {
		return
	}
	l.checkFile()
	f := "[DEBUG] " + format
	l.Printf(f, v...)
}

func (l *Logs) Info(format string, v ...interface{}) {
	if l.level > InfoLevel {
		return
	}
	l.checkFile()
	f := "[INFO] " + format
	l.Printf(f, v...)
}

func (l *Logs) Warn(format string, v ...interface{}) {
	if l.level > WarnLevel {
		return
	}
	l.checkFile()
	f := "[WARN] " + format
	l.Printf(f, v...)
}

func (l *Logs) Error(format string, v ...interface{}) {
	if l.level > ErrorLevel {
		return
	}
	l.checkFile()
	f := "[ERROR] " + format
	l.Printf(f, v...)
}

func (l *Logs) Fatal(format string, v ...interface{}) {
	l.checkFile()
	f := "[FATAL] " + format
	l.Fatalf(f, v...)
}

func (l *Logs) Panic(format string, v ...interface{}) {
	l.checkFile()
	f := "[PANIC] " + format
	l.Panicf(f, v...)
}

//关闭日志文件
func (l *Logs) CloseFile() {
	if l.f != nil {
		l.f.Close()
	}
}

func New(logger *log.Logger, lev Level) *Logs {
	return &Logs{logger, lev, "", nil}
}

func SetLevel(level Level) {
	std.SetLevel(level)
}

func ParseLevel(level string) {
	std.ParseLevel(level)
}

func GetLevel() string {
	return std.GetLevel()
}

//设置输出
func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

//设置日志文件名称
func SetLogFile(name string) {
	std.SetLogFile(name)
}

func Debug(format string, v ...interface{}) {
	std.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	std.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	std.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	std.Error(format, v...)
}

func Fatal(format string, v ...interface{}) {
	std.Fatal(format, v...)
}

func Panic(format string, v ...interface{}) {
	std.Panic(format, v...)
}

func CloseFile() {
	std.CloseFile()
}

var std *Logs

func init() {
	// Defautl logger & functions
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	//fmt.Println(file)
	_, filename := path.Split(file)
	prefix := fmt.Sprintf("[%s:%d] ", filename, line)
	std = New(log.New(os.Stderr, prefix, log.Ldate|log.Ltime), DebugLevel)
}
