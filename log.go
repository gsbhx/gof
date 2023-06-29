/**
 *
 * @Author wangkan
 * @Date   2021/1/26 下午1:29
 */
package gof

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Logger struct {
	debugLogger *loggerObject
	infoLogger  *loggerObject
	errorLogger *loggerObject
	fatalLogger *loggerObject
}

type loggerObject struct {
	file     *os.File
	obj      *log.Logger
	lastDate *time.Time
	mu       *sync.RWMutex
}

const (
	DATEFORMATE     = "2006-01-02"
	DEBUG       int = iota
	INFO
	ERROR
	FATAL
)

var Log *Logger
var isConsole = true
var level = 1
var logDir = ""

func (l *Logger) InitLogger(dir string) (err error) {
	if err = createDir(dir); err != nil {
		fmt.Println("mkdir dir failed", err.Error())
		return
	}
	l.infoLogger = new(loggerObject)
	l.debugLogger = new(loggerObject)
	l.errorLogger = new(loggerObject)
	l.fatalLogger = new(loggerObject)
	makeLoggerObj(l.infoLogger, "INFO")
	makeLoggerObj(l.debugLogger, "DEBUG")
	makeLoggerObj(l.errorLogger, "ERROR")
	makeLoggerObj(l.fatalLogger, "FATAL")
	return
}

func (l *Logger) Info(format string, args ...interface{}) {
	if level <= INFO {
		l.infoLogger.write("INFO", format, args...)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if level <= DEBUG {
		l.debugLogger.write("DEBUG", format, args...)
	}
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	if level <= FATAL {
		l.fatalLogger.write("FATAL", format, args...)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if level <= ERROR {
		l.errorLogger.write("ERROR", format, args...)
	}
}
func (l *Logger) SetLevel(userLevel int) {
	level = userLevel
}

func (l *Logger) SetConsole(console bool) {
	isConsole = console
}

func createDir(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			if os.IsPermission(err) {
				fmt.Println("create dir error:", err.Error())
				e = err
			}
		}
	}
	return
}

func (l *loggerObject) write(levelString string, format string, args ...interface{}) {
	isNewDay := l.isNewDay()
	if isNewDay {
		makeLoggerObj(l, levelString)
	}
	str := fmt.Sprintf(format, args...)
	if l.obj != nil {
		l.obj.Println(str)
	}
	if isConsole {
		fmt.Printf("%+v\n", str)
	}

}

func makeLoggerObj(l *loggerObject, name string) {
	now := time.Now().Format(DATEFORMATE)
	l.mu = new(sync.RWMutex)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		_ = l.file.Close()
	}

	t, _ := time.Parse(DATEFORMATE, now)
	l.lastDate = &t
	//fileName := logDir + "/" + name + "-" + now + ".log"
	fileName := logDir + "/" + now + ".log"
	f, _err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)

	if _err == nil {
		l.file = f
		l.obj = log.New(l.file, "["+name+"]  ", log.Ldate|log.Ltime)
	}
}

func (l *loggerObject) isNewDay() bool {
	now := time.Now().Format(DATEFORMATE)
	t, _ := time.Parse(DATEFORMATE, now)
	return t.After(*l.lastDate)
}

func init() {
	Log = new(Logger)
	Log.SetConsole(false)
	_ = Log.InitLogger("Logs")
}
