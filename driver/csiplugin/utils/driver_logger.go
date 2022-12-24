package utils

import (
	"flag"
	"os"
)


const logLevel = "LOGLEVEL"

type LoggerLevel int

const (
	DEBUGPLUS LoggerLevel = iota
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
)

func (level LoggerLevel) String() string {
	switch level {
	case DEBUGPLUS:
		return "DEBUGPLUS"
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
	if level == "" || level == DEBUG.String() || level == DEBUGPLUS.String(){
		logValue = INFO.String()
	}else{
		logValue = level
	}
	_ = flag.Set("alsologtostderr", "false")
	_ = flag.Set("stderrthreshold", logValue)
	if level == DEBUG.String() {
		_ = flag.Set("v", "4")
	} else if level == DEBUGPLUS.String() {
		_ = flag.Set("v", "6")
	} else {
		_ = flag.Set("v", "2")
	}
	_ = flag.Set("log_dir", "/host/var/log/")
	flag.Parse()
}

