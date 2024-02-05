package logger

import (
	"fmt"
	"time"
)

type Logger struct {
	LogId       string
	Debug       bool
	TimeCreated time.Time
}

// Public methods
func (l *Logger) LogInfo(msg string) {
	l.writeToLog("INFO", msg)
}

func (l *Logger) LogError(msg string) {
	l.writeToLog("ERROR", msg)
}

func (l *Logger) LogDebug(msg string) {
	if l.Debug {
		l.writeToLog("DEBUG", msg)
	}
}

// Private methods
func (l *Logger) writeToLog(severity string, msg string) {
	if l.LogId == "" {
		l.LogId = "no-id"
	}
	fmt.Printf("%s %s +%s [id=%s] %s\n",
		time.Now().UTC().Format("2006/01/02 15:04:05"),
		severity,
		l.sinceCreated(),
		l.LogId,
		msg)
}

func (l *Logger) sinceCreated() string {
	t := time.Since(l.TimeCreated).Seconds()
	return fmt.Sprintf("%.1fs", t)
}
