package zlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/codeslala/igo/env"
	"github.com/codeslala/igo/util"
)

const (
	timeFormat     = "2006-01-02 15:04:05.000"
	filenameFormat = "20060102"
)

var line = func() int {
	if os.Getenv(env.GoEnv) == env.ProductMode {
		return 100
	}
	return 5
}()

type LogWriter struct {
	logDir string

	level       string
	curFilename string
	writer      *FileWriter
	mu          sync.Mutex
	buffer      []string

	timeInterval time.Duration
	ticker       *time.Ticker
}

func (l *LogWriter) Write(close bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.write(close)
}

func (l *LogWriter) filename(t time.Time) string {
	return filepath.Join(l.logDir, t.Format(filenameFormat)+".log")
}

func (l *LogWriter) rotate(t time.Time) error {
	l.write(true)
	name := l.filename(t)
	l.writer = &FileWriter{File: name}
	l.curFilename = name
	return nil
}

func (l *LogWriter) write(close bool) {
	if len(l.buffer) == 0 {
		return
	}
	tmpBuf := l.buffer
	asyncWriteFile(l.writer, tmpBuf, close)
	l.buffer = make([]string, 0, line)
}

func (l *LogWriter) startTicker() {
	if l.timeInterval == 0 {
		return
	}
	if l.ticker != nil {
		return
	}
	l.ticker = time.NewTicker(l.timeInterval)
	go func() {
		for {
			select {
			case <-l.ticker.C:
				l.Write(false)
			}
		}
	}()
}

type Config struct {
	Info  *LogWriter
	Error *LogWriter
	Trace *LogWriter
}

func (c *Config) lockAndFlushAll() {
	if c.Info != nil {
		l := c.Info
		l.mu.Lock()
		defer l.mu.Unlock()
		if len(c.Info.buffer) > 0 {
			tmpBuf := l.buffer
			l.buffer = make([]string, 0, line)
			syncWriteFile(l.writer, tmpBuf)
		}
	}
	if c.Error != nil {
		l := c.Error
		l.mu.Lock()
		defer l.mu.Unlock()
		if len(c.Error.buffer) > 0 {
			tmpBuf := l.buffer
			l.buffer = make([]string, 0, line)
			syncWriteFile(l.writer, tmpBuf)
		}
	}
	if c.Trace != nil {
		l := c.Trace
		l.mu.Lock()
		defer l.mu.Unlock()
		if len(c.Trace.buffer) > 0 {
			tmpBuf := l.buffer
			l.buffer = make([]string, 0, line)
			syncWriteFile(l.writer, tmpBuf)
		}
	}
}

var std *Config

func Info(s string, sync bool) {
	write(std.Info, s, sync)
}

func Error(err error, sync bool) {
	write(std.Error, fmt.Sprintf("%v", err), sync)
}

func Trace(s string, sync bool) {
	write(std.Trace, s, sync)
}

func write(l *LogWriter, s string, sync bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	t := time.Now()
	log := s
	if l.level != "" {
		log = t.Format(timeFormat) + " [" + l.level + "] " + log
	}
	if l.writer == nil {
		l.writer = &FileWriter{File: l.filename(t)}
		l.curFilename = l.filename(t)
	}
	if l.filename(t) != l.curFilename {
		l.rotate(t)
		if !sync {
			l.buffer = append(l.buffer, log)
			return
		}
	}
	if sync {
		syncWriteFile(l.writer, []string{log})
		return
	}
	l.startTicker()
	if len(l.buffer) < line {
		l.buffer = append(l.buffer, log)
	}
	if len(l.buffer) < line {
		return
	}
	l.write(false)
	return
}

func asyncWriteFile(fileWriter *FileWriter, buf []string, close bool) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				syncWriteFile(std.Error.writer, []string{fmt.Sprintf("%v", err)})
			}
		}()
		syncWriteFile(fileWriter, buf)
		if close {
			fileWriter.Close()
		}
	}()
}

func syncWriteFile(fileWriter *FileWriter, buf []string) {
	if fileWriter != nil {
		fileWriter.Write(util.StringToBytes(strings.Join(buf, "\n") + "\n"))
	}
}

func init() {
	std = &Config{
		Info: &LogWriter{
			timeInterval: 10 * time.Minute,
			level:        "info",
			logDir:       "info",
			buffer:       make([]string, 0, line),
		},
		Error: &LogWriter{
			timeInterval: 10 * time.Minute,
			level:        "error",
			logDir:       "error",
			buffer:       make([]string, 0, line),
		},
		Trace: &LogWriter{
			timeInterval: 10 * time.Minute,
			level:        "trace",
			logDir:       "trace",
			buffer:       make([]string, 0, line),
		},
	}
}
