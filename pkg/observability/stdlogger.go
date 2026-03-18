package observability

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
)

// StdLogger implements Logger using log.Logger and emits key=value pairs.
type StdLogger struct {
	log  *log.Logger
	pre  []any // prefix keyvals
}

// NewStdLogger returns a Logger that writes to the given log.Logger.
// Keyvals must be alternating key, value; keys are converted to strings.
func NewStdLogger(l *log.Logger) *StdLogger {
	if l == nil {
		l = log.New(os.Stdout, "", log.LstdFlags|log.LUTC)
	}
	return &StdLogger{log: l, pre: nil}
}

func (s *StdLogger) appendKeyvals(buf *bytes.Buffer, keyvals []any) {
	for i := 0; i+1 < len(keyvals); i += 2 {
		if buf.Len() > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(formatKey(keyvals[i]))
		buf.WriteByte('=')
		buf.WriteString(formatValue(keyvals[i+1]))
	}
}

func formatKey(k any) string {
	switch v := k.(type) {
	case string:
		return v
	default:
		return "key"
	}
}

func formatValue(v any) string {
	switch x := v.(type) {
	case string:
		if strings.ContainsAny(x, " \t\n\"") {
			return `"` + strings.ReplaceAll(x, `"`, `\"`) + `"`
		}
		return x
	default:
		s := fmt.Sprint(v)
		if strings.ContainsAny(s, " \t\n\"") {
			return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
		}
		return s
	}
}

func (s *StdLogger) Info(msg string, keyvals ...any) {
	var buf bytes.Buffer
	buf.WriteString("level=info ")
	buf.WriteString("msg=")
	buf.WriteString(formatValue(msg))
	if len(s.pre) > 0 {
		s.appendKeyvals(&buf, s.pre)
	}
	if len(keyvals) > 0 {
		s.appendKeyvals(&buf, keyvals)
	}
	s.log.Output(2, buf.String())
}

func (s *StdLogger) Warn(msg string, keyvals ...any) {
	var buf bytes.Buffer
	buf.WriteString("level=warn ")
	buf.WriteString("msg=")
	buf.WriteString(formatValue(msg))
	if len(s.pre) > 0 {
		s.appendKeyvals(&buf, s.pre)
	}
	if len(keyvals) > 0 {
		s.appendKeyvals(&buf, keyvals)
	}
	s.log.Output(2, buf.String())
}

func (s *StdLogger) Error(msg string, keyvals ...any) {
	var buf bytes.Buffer
	buf.WriteString("level=error ")
	buf.WriteString("msg=")
	buf.WriteString(formatValue(msg))
	if len(s.pre) > 0 {
		s.appendKeyvals(&buf, s.pre)
	}
	if len(keyvals) > 0 {
		s.appendKeyvals(&buf, keyvals)
	}
	s.log.Output(2, buf.String())
}

func (s *StdLogger) With(keyvals ...any) Logger {
	pre := make([]any, 0, len(s.pre)+len(keyvals))
	pre = append(pre, s.pre...)
	pre = append(pre, keyvals...)
	return &StdLogger{log: s.log, pre: pre}
}
