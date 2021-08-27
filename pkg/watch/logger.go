package watch

import (
	"fmt"
	"io"
	"log"
)

type Logger interface {
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func NewLogger(w io.Writer, debug bool) Logger {
	return &stdLogger{
		l:     log.New(w, "", log.LstdFlags),
		debug: debug,
	}
}

type stdLogger struct {
	l     *log.Logger
	debug bool
}

func (s *stdLogger) Infof(format string, args ...interface{}) {
	s.l.Printf("[INFO]: %s", fmt.Sprintf(format, args...))
}

func (s *stdLogger) Debugf(format string, args ...interface{}) {
	if s.debug {
		s.l.Printf("[DEBUG]: %s", fmt.Sprintf(format, args...))
	}
}

func (s *stdLogger) Errorf(format string, args ...interface{}) {
	s.l.Printf("[ERROR]: %s", fmt.Sprintf(format, args...))
}
