package utils

import (
	"flag"
	"os"

	"github.com/golang/glog"
)

const logLevel = "LOGLEVEL"
const path = "scalecsilogs"

type LoggerLevel int

const (
	TRACE LoggerLevel = iota
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
)

func (level LoggerLevel) String() string {
	switch level {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	case INFO:
		return "INFO"
	default:
		return "INFO"
	}
}

func InitLogger() {
	level := os.Getenv(logLevel)
	var logValue string
	if level == "" || level == DEBUG.String() || level == TRACE.String() {
		logValue = INFO.String()
	} else {
		logValue = level
	}
	_ = flag.Set("alsologtostderr", "false")
	_ = flag.Set("stderrthreshold", logValue)
	glog.Infof("logValue: %s", logValue)
	if level == DEBUG.String() {
		_ = flag.Set("v", "4")
	} else if level == TRACE.String() {
		_ = flag.Set("v", "6")
	} else {
		_ = flag.Set("v", "0")
	}

	dirPath := "/host/var/log/" + path + "/"
	if !Exists(dirPath) {
		err := MkDir(dirPath)
		if err != nil {
			glog.Errorf("Failed to create log directory")
		}
	}

	_ = flag.Set("log_dir", dirPath)
	flag.Parse()
}
