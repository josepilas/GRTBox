package logs

import (
	"sync"
	"time"
)

type Entry struct {
	ID        int64  `json:"id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type Logger struct {
	mu      sync.Mutex
	nextID  int64
	entries []Entry
}

func NewLogger() *Logger {
	return &Logger{
		nextID:  1,
		entries: make([]Entry, 0, 128),
	}
}

func (l *Logger) Info(message string) Entry {
	return l.Add("info", message)
}

func (l *Logger) Warn(message string) Entry {
	return l.Add("warn", message)
}

func (l *Logger) Error(message string) Entry {
	return l.Add("error", message)
}

func (l *Logger) Add(level string, message string) Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := Entry{
		ID:        l.nextID,
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
	}
	l.nextID++
	l.entries = append(l.entries, entry)

	if len(l.entries) > 500 {
		l.entries = l.entries[len(l.entries)-500:]
	}

	return entry
}

func (l *Logger) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}
