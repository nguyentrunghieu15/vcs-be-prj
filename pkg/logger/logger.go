package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type LevelLogType int

const (
	INFO LevelLogType = iota
	WARN
	ERROR
	DEBUG
)

type LoggerConfig struct {
	IsLogRotate     bool
	PathToLog       string
	FileNameLogBase string
}

type LoggerDecorator struct {
	_today time.Time
	l      *log.Logger
	mu     sync.Mutex
	Config LoggerConfig
}

func (logger *LoggerDecorator) SetOutput(w io.Writer) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.l.SetOutput(w)
}

func (logger *LoggerDecorator) SetToday(today time.Time) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger._today = today
}

func (logger *LoggerDecorator) Log(level LevelLogType, fields map[string]interface{}) {
	if !dateEqual(logger._today, time.Now()) && logger.Config.IsLogRotate {
		logger.SetToday(time.Now())

		dailyLogFile := logger.Config.PathToLog + logger.Config.FileNameLogBase + "-" + time.Now().Format("2006-01-02") + ".log"
		newF, err := os.OpenFile(dailyLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Panicf("error opening file: %v", err)
		}
		wr := io.MultiWriter(newF, os.Stdout)
		logger.SetOutput(wr)
	}
	logMessage := ""
	switch level {
	case INFO:
		logMessage += "INFO |"
	case WARN:
		logMessage += "WARN |"
	case ERROR:
		logMessage += "ERROR |"
	case DEBUG:
		logMessage += "DEBUG |"
	}

	logMessage += fmt.Sprintf("time: %v |", time.Now())

	for k, v := range fields {
		logMessage += fmt.Sprintf("%v: %v |", k, v)
	}

	logMessage = strings.TrimSuffix(logMessage, " |")
	logger.l.Println(logMessage)
}

func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func NewLogger() *LoggerDecorator {
	return &LoggerDecorator{l: &log.Logger{}}
}

func (logger *LoggerDecorator) ImplementedMiddlewareLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !dateEqual(logger._today, time.Now()) && logger.Config.IsLogRotate {
				logger.SetToday(time.Now())

				dailyLogFile := logger.Config.PathToLog + logger.Config.FileNameLogBase + "-" + time.Now().Format("2006-01-02") + ".log"
				newF, err := os.OpenFile(dailyLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
				if err != nil {
					log.Panicf("error opening file: %v", err)
				}
				wr := io.MultiWriter(newF, os.Stdout)
				logger.SetOutput(wr)
			}

			err := next(c)
			if err != nil {
				c.Error(err)
				logger.l.Printf("Error | Time: %v | Uri: %v | Method: %v | Status: %v | RemoteAddr: %v | Error:%v \n",
					time.Now(),
					c.Request().RequestURI,
					c.Request().Method,
					c.Response().Status,
					c.Request().RemoteAddr,
					err,
				)
			}

			logger.l.Printf("INFO | Time: %v | Uri: %v | Method: %v | Status: %v | RemoteAddr: %v \n",
				time.Now(),
				c.Request().RequestURI,
				c.Request().Method,
				c.Response().Status,
				c.Request().RemoteAddr,
			)
			return err
		}
	}
}
