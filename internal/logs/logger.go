package logs

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	ioz "github.com/kunitsucom/util.go/io"
)

//nolint:gochecknoglobals
var (
	Trace Logger = NewDiscard() //nolint:revive
	Debug Logger = NewDiscard() //nolint:revive
	Info  Logger = logger{log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)}
	Warn  Logger = logger{log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)}
)

func NewDiscard() Logger { //nolint:ireturn
	return logger{log.New(io.Discard, "", 0)}
}

func NewTrace() Logger { //nolint:ireturn
	return logger{log.New(os.Stderr, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)}
}

func NewDebug() Logger { //nolint:ireturn
	return logger{log.New(os.Stderr, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)}
}

type Logger interface {
	io.Writer
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	LineWriter(prefix string) io.Writer
}

const callerSkip = 2

type logger struct{ l *log.Logger }

func (l logger) Print(v ...interface{}) { _ = l.l.Output(callerSkip, fmt.Sprint(v...)) }
func (l logger) Printf(format string, v ...interface{}) {
	_ = l.l.Output(callerSkip, fmt.Sprintf(format, v...))
}

func (l logger) Write(p []byte) (n int, err error) {
	l.Print(string(p))
	return len(p), nil
}

func (l logger) LineWriter(prefix string) io.Writer {
	return ioz.WriteFunc(func(p []byte) (n int, err error) {
		lines := bytes.Split(p, []byte("\n"))
		for _, line := range lines {
			_ = l.l.Output(1, prefix+string(line))
		}

		return len(p), nil
	})
}
