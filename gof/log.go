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
	DATEFORMATE = "2006-01-02"
	DEBUG int = iota
	INFO
	ERROR
	FATAL
)

var Log *Logger
var isConsole = true
var level = 1
var logDir = ""

func (this *Logger) InitLogger(dir string) (e error) {
	logDir = dir
	err := createDir(logDir)
	if err != nil {
		fmt.Println("mkdir dir failed")
		e = err
	} else {
		this.infoLogger = new(loggerObject)
		this.debugLogger = new(loggerObject)
		this.errorLogger = new(loggerObject)
		this.fatalLogger = new(loggerObject)

		makeLoggerObj(this.infoLogger, "INFO")
		makeLoggerObj(this.debugLogger, "DEBUG")
		makeLoggerObj(this.errorLogger, "ERROR")
		makeLoggerObj(this.fatalLogger, "FATAL")
	}
	return
}

func (this *Logger) Info(format string, args ...interface{}) {
	if level <= INFO {
		this.infoLogger.write("INFO", format, args...)
	}
}

func (this *Logger) Debug(format string, args ...interface{}) {
	if level <= DEBUG {
		this.debugLogger.write("DEBUG", format, args...)
	}
}

func (this *Logger) Fatal(format string, args ...interface{}) {
	if level <= FATAL {
		this.fatalLogger.write("FATAL", format, args...)
	}
}

func (this *Logger) Error(format string, args ...interface{}) {
	if level <= ERROR {
		this.errorLogger.write("ERROR", format, args...)
	}
}
func (this *Logger) SetLevel(userLevel int) {
	level = userLevel
}

func (this *Logger) SetConsole(console bool) {
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

func (this *loggerObject) write(levelString string, format string, args ...interface{}) {
	isNewDay := this.isNewDay()
	if isNewDay {
		makeLoggerObj(this, levelString)
	}
	str := fmt.Sprintf(format, args...)
	if this.obj != nil {
		this.obj.Println(str)
	}
	if isConsole {
		fmt.Println(str)
	}
}

func makeLoggerObj(l *loggerObject, name string) {
	now := time.Now().Format(DATEFORMATE)
	l.mu = new(sync.RWMutex)
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
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

func (this *loggerObject) isNewDay() bool {
	now := time.Now().Format(DATEFORMATE)
	t, _ := time.Parse(DATEFORMATE, now)
	return t.After(*this.lastDate)
}

func init()  {
	Log =new(Logger)
	Log.InitLogger("Logs")
}
