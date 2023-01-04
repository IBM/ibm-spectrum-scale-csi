package utils

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/klog/v2"
	"os"
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

func InitLogger(ctx context.Context) {
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
		_ = flag.Set("v", "1")
	}

	dirPath := "/host/var/log/" + path + "/"
	if !Exists(ctx, dirPath) {
		err := MkDir(ctx, dirPath)
		if err != nil {
			glog.Errorf("Failed to create log directory")
		}
	}

	_ = flag.Set("log_dir", dirPath)
	flag.Parse()
}

func init() {
	klog.NewKlogr()
}

type Logger interface {
	Trace(ctx context.Context, format string, args ...interface{})
}

type CsiLogger struct{}

func (l *CsiLogger) Trace(ctx context.Context, format string, args ...interface{}) {
	arg := setFormat(ctx, format)
	klog.V(6).InfofDepth(1, format, arg)
}

func setFormat(ctx context.Context, format string) string {
	var logFormat string
	if ctx != nil {
		loggerId := GetLoggerId(ctx)
		logFormat = fmt.Sprintf("[%s] %s", loggerId, format)
		return logFormat
	} else {
		return format
	}
}
