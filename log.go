package kproxy

import (
	"log"
	"os"
	"sync"
)

var (
	logger        = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	loggerMu      sync.Mutex
	enableLogging bool
)

// SetLogger sets a logger and enable logging.
func SetLogger(l *log.Logger) {
	if l == nil {
		return
	}

	loggerMu.Lock()
	logger = l
	EnableLogging()
	loggerMu.Unlock()
}

// EnableLogging enables logging with a default logger.
func EnableLogging() {
	enableLogging = true
}

func klog(format string, args ...interface{}) {
	if enableLogging {
		log.Printf(format, args...)
	}
}
