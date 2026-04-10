// Package audit provides a multi-sink audit trail for all tool invocations.
package audit

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"time"
)

// Event is an immutable audit record for every tool invocation.
type Event struct {
	EventID        string         `json:"event_id"`
	Timestamp      time.Time      `json:"timestamp"`
	Interface      string         `json:"interface"`
	User           string         `json:"user"`
	Tool           string         `json:"tool"`
	Target         string         `json:"target"`
	Params         map[string]any `json:"parameters,omitempty"`
	ConfirmType    string         `json:"confirm_type,omitempty"`
	ConfirmEcho    string         `json:"confirm_echo,omitempty"`
	Result         string         `json:"result,omitempty"`
	Error          string         `json:"error,omitempty"`
	DurationMs     int64          `json:"duration_ms,omitempty"`
	RedactedFields []string       `json:"redacted_fields,omitempty"`

	startTime time.Time
}

// NewEvent creates an audit event with a unique ID and timestamp.
func NewEvent(iface, user, tool, target string) *Event {
	return &Event{
		EventID:   generateID(),
		Timestamp: time.Now(),
		Interface: iface,
		User:      user,
		Tool:      tool,
		Target:    target,
		startTime: time.Now(),
	}
}

// Complete finalizes the event with a result and optional error.
func (e *Event) Complete(result string, err error) {
	e.Result = result
	if err != nil {
		e.Error = err.Error()
	}
	e.DurationMs = time.Since(e.startTime).Milliseconds()
}

// Redact replaces sensitive parameter values with a placeholder.
func (e *Event) Redact(fields []string) {
	if e.Params == nil {
		return
	}
	for _, f := range fields {
		if _, ok := e.Params[f]; ok {
			e.Params[f] = "***REDACTED***"
			e.RedactedFields = append(e.RedactedFields, f)
		}
	}
}

// Sink writes audit events to a destination.
type Sink func(*Event)

// Logger writes audit events to one or more sinks.
type Logger struct {
	sinks []Sink
}

// LoggerOption configures the Logger.
type LoggerOption func(*Logger)

// WithStdoutSink adds a JSON-line sink writing to w.
func WithStdoutSink(w io.Writer) LoggerOption {
	return func(l *Logger) {
		l.sinks = append(l.sinks, func(e *Event) {
			data, _ := json.Marshal(e)
			w.Write(data)
			w.Write([]byte("\n"))
		})
	}
}

// NewLogger creates a multi-sink audit logger.
func NewLogger(opts ...LoggerOption) *Logger {
	l := &Logger{}
	for _, o := range opts {
		o(l)
	}
	return l
}

// Log writes an event to all sinks.
func (l *Logger) Log(e *Event) {
	for _, s := range l.sinks {
		s(e)
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
