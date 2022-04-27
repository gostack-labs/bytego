package bytego

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

const (
	LEVEL_DEBUG LogLevel = iota + 1
	LEVEL_INFO
	LEVEL_WARN
	LEVEL_ERROR
	LEVEL_OFF
)

type LogLevel int

func (l LogLevel) Int() int {
	return int(l)
}

func (l LogLevel) String() string {
	switch l {
	case LEVEL_DEBUG:
		return "DEBUG"
	case LEVEL_INFO:
		return "INFO"
	case LEVEL_WARN:
		return "WARN"
	case LEVEL_ERROR:
		return "ERROR"
	case LEVEL_OFF:
		return "OFF"
	default:
		return ""
	}
}

func ParseLogLevel(s string) LogLevel {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LEVEL_DEBUG
	case "INFO":
		return LEVEL_INFO
	case "WARN":
		return LEVEL_WARN
	case "ERROR":
		return LEVEL_ERROR
	case "OFF":
		return LEVEL_OFF
	}
	return LEVEL_INFO
}

type Logger interface {
	SetLevel(level string)

	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})

	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

func NewLogger(w io.Writer, level ...string) Logger {
	var logLevel LogLevel = LEVEL_INFO
	if len(level) > 0 {
		logLevel = ParseLogLevel(level[0])
	}
	var l Logger = &defaultLogger{
		level: logLevel,
		log:   log.New(w, "", log.LstdFlags),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
	return l
}

type defaultLogger struct {
	log   *log.Logger
	level LogLevel
	pool  *sync.Pool
}

func (l *defaultLogger) SetLevel(level string) {
	l.level = ParseLogLevel(level)
}

func (l *defaultLogger) Debug(v ...interface{}) {
	_ = l.write(LEVEL_DEBUG, fmt.Sprint(v...))
}

func (l *defaultLogger) Info(v ...interface{}) {
	_ = l.write(LEVEL_INFO, fmt.Sprint(v...))
}

func (l *defaultLogger) Warn(v ...interface{}) {
	_ = l.write(LEVEL_WARN, fmt.Sprint(v...))
}

func (l *defaultLogger) Error(v ...interface{}) {
	_ = l.write(LEVEL_ERROR, fmt.Sprint(v...))
}

func (l *defaultLogger) Debugf(format string, v ...interface{}) {
	_ = l.write(LEVEL_DEBUG, format, v...)
}

func (l *defaultLogger) Infof(format string, v ...interface{}) {
	_ = l.write(LEVEL_INFO, format, v...)
}

func (l *defaultLogger) Warnf(format string, v ...interface{}) {
	_ = l.write(LEVEL_WARN, format, v...)
}

func (l *defaultLogger) Errorf(format string, v ...interface{}) {
	_ = l.write(LEVEL_ERROR, format, v...)
}

func (l *defaultLogger) write(level LogLevel, format string, v ...interface{}) error {
	if level < l.level {
		return nil
	}
	buf := l.pool.Get().(*bytes.Buffer)
	buf.WriteString(l.level.String())
	buf.WriteString(" ")
	_, _ = fmt.Fprintf(buf, format, v...)
	_ = l.log.Output(4, buf.String())
	buf.Reset()
	l.pool.Put(buf)
	return nil
}
