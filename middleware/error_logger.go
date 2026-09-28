package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	errorLogFile *os.File
	errorLogLock sync.Mutex
	errorLogger  *log.Logger
)

// InitErrorLogger initializes the error.log file and sets up the file writer.
func InitErrorLogger(logFilePath string) (*os.File, error) {
	if logFilePath == "" {
		logFilePath = "error.log"
	}

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("[Error] Failed to open error log file '%s': %v", logFilePath, err)
		return nil, err
	}

	errorLogFile = file
	errorLogger = log.New(io.MultiWriter(os.Stderr, file), "", log.LstdFlags)

	// Log initial start marker
	WriteErrorLog("=== Server Error Logger Initialized at " + time.Now().Format("2006-01-02 15:04:05") + " ===")
	return file, nil
}

// WriteErrorLog writes a formatted message to error.log thread-safely.
func WriteErrorLog(message string) {
	errorLogLock.Lock()
	defer errorLogLock.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s] %s\n", timestamp, message)

	if errorLogFile != nil {
		_, _ = errorLogFile.WriteString(entry)
		_ = errorLogFile.Sync()
	}
}

// responseBodyWriter captures response body for logging error messages
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// ErrorLoggerMiddleware intercepts HTTP requests and records any 4xx / 5xx responses or panics to error.log.
func ErrorLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()

		wb := &responseBodyWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = wb

		// Defer panic recovery specifically for logging detailed crash to error.log
		defer func() {
			if r := recover(); r != nil {
				latency := time.Since(start)
				errStr := fmt.Sprintf("PANIC RECOVERED: %v | Latency: %v | Method: %s | Path: %s | ClientIP: %s",
					r, latency, method, path, clientIP)
				WriteErrorLog("[CRITICAL] " + errStr)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Internal server error occurred (recovered from panic)",
				})
			}
		}()

		c.Next()

		statusCode := c.Writer.Status()
		latency := time.Since(start)

		if statusCode >= 400 {
			fullPath := path
			if rawQuery != "" {
				fullPath += "?" + rawQuery
			}

			respBody := wb.body.String()
			if len(respBody) > 500 {
				respBody = respBody[:500] + "... (truncated)"
			}

			logLevel := "WARNING"
			if statusCode >= 500 {
				logLevel = "ERROR"
			}

			logMsg := fmt.Sprintf("[%s] HTTP %d | Method: %-6s | Path: %s | Client: %s | Latency: %v | Response: %s",
				logLevel, statusCode, method, fullPath, clientIP, latency, respBody)

			WriteErrorLog(logMsg)
		}
	}
}
