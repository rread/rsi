package log

import (
	"fmt"
	"os"
)
import l "log"

type Level int
type Loggerf func(string, ...interface{})
type Loggerln func(...interface{})

const (
	FatalLevel = Level(iota)
	Error
	Info
	Debug
)

var (
	level  Level = Info
	logger *l.Logger

	Debugf  = GenLoggerf(Debug)
	Debugln = GenLoggerln(Debug)
	Printf  = GenLoggerf(Debug)
	Println = GenLoggerln(Debug)
	Infof   = GenLoggerf(Info)
	Infoln  = GenLoggerln(Info)
	Errorf  = GenLoggerf(Error)
	Errorln = GenLoggerln(Error)
	Fatalf  = GenLoggerf(FatalLevel)
	Fatal   = GenLoggerln(FatalLevel)
)

func init() {
	logger = l.New(os.Stderr, "", l.LstdFlags|l.Lshortfile)
}

func SetLevel(l Level) {
	level = l
}

func GenLoggerf(l Level) Loggerf {
	n := l
	return func(format string, v ...interface{}) {
		if level >= n {
			logger.Output(2, fmt.Sprintf(format, v...))
		}
		if n == FatalLevel {
			os.Exit(1)
		}
	}
}

func GenLoggerln(l Level) Loggerln {
	n := l
	return func(v ...interface{}) {
		if level >= n {
			logger.Output(2, fmt.Sprint(v...))
		}
		if n == FatalLevel {
			os.Exit(1)
		}
	}
}
