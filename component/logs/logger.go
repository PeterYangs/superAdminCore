package logs

import (
	"fmt"
	"github.com/PeterYangs/tools"
	"github.com/spf13/cast"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type message struct {
	message string
	level   level
}

func NewMessage(m string, l level) *message {

	return &message{message: m, level: l}
}

func (m *message) Println() {

	fmt.Println(m.message)

}

func (m *message) Print() {

	fmt.Print(m.message)
}

func (m *message) Message() string {

	return m.message
}

type logger struct {
	lock      sync.Mutex
	dirConfig map[level]string
}

type level string

func (l level) ToString() string {

	return string(l)
}

const (
	Error level = "ERROR"
	Info  level = "INFO"
	Debug level = "DEBUG"
)

var Logger *logger

var once sync.Once

func NewLogger() *logger {

	once.Do(func() {

		dir := map[level]string{
			Info:  "logs/info",
			Error: "logs/error",
			Debug: "logs/debug",
		}

		for _, s := range dir {

			_ = os.MkdirAll(s, 0755)
		}

		Logger = &logger{dirConfig: dir, lock: sync.Mutex{}}

	})

	return Logger
}

func (l *logger) Info(message string) *message {

	return NewMessage(l.common(Info, message), Info)
}

func (l *logger) Debug(message string) *message {

	return NewMessage(l.common(Debug, message), Debug)
}

func (l *logger) Error(message string) *message {

	return NewMessage(l.common(Error, message), Error)
}

//-----------------------------------------------------------------------------------------------------------------

func (l *logger) common(lv level, message string) string {

	f, e := os.OpenFile(l.getDirByLevel(lv)+"/"+tools.Date("Y-m-d", time.Now().Unix())+".log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)

	if e != nil {

		return ""
	}

	defer f.Close()

	m := l.logFormat(lv, message)

	f.Write([]byte(m))

	return m

}

func (l *logger) getDirByLevel(lv level) string {

	l.lock.Lock()

	defer l.lock.Unlock()

	return l.dirConfig[lv]

}

func (l *logger) logFormat(level level, message string) string {

	_, f, lr, _ := runtime.Caller(3)

	m := "[" + level.ToString() + "] " + tools.Date("Y-m-d H:i:s", time.Now().Unix()) + "  " + message + "\n\t" + f + ":" + cast.ToString(lr) + " " + "\n"

	if level == Error {

		m += strings.Replace("\n[stacktrace]\n"+string(debug.Stack())+"\n\n", "\n", "\n\t", -1)
	}

	return m

}
