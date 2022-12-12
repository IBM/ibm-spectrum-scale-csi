package utils

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/glog"
)

type CsiLogger struct{}

var logger *CsiLogger

func (log *CsiLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	if ctx == nil {
		glog.Infof(format, args...)
	} else {
		loggerId := GetLoggerId(ctx)
		logFormat := fmt.Sprintf("[%s] %s", loggerId, format)
		glog.Infof(logFormat, args)
	}
}

func (log *CsiLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	if ctx == nil {
		glog.Infof(format, args...)
	} else {
		loggerId := GetLoggerId(ctx)
		logFormat := fmt.Sprintf("[%s] %s", loggerId, format)
		glog.V(4).Infof(logFormat, args)
	}
}

func (log *CsiLogger) DebugPlus(ctx context.Context, format string, args ...interface{}) {
	if ctx == nil {
		glog.Infof(format, args...)
	} else {
		loggerId := GetLoggerId(ctx)
		logFormat := fmt.Sprintf("[%s] %s", loggerId, format)
		glog.V(6).Infof(logFormat, args)
	}
}

func (log *CsiLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	if ctx == nil {
		glog.Infof(format, args...)
	} else {
		loggerId := GetLoggerId(ctx)
		logFormat := fmt.Sprintf("[%s] %s", loggerId, format)
		glog.Errorf(logFormat, args)
	}
}

func (log *CsiLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	if ctx == nil {
		glog.Infof(format, args...)
	} else {
		loggerId := GetLoggerId(ctx)
		logFormat := fmt.Sprintf("[%s] %s", loggerId, format)
		glog.Fatalf(logFormat, args)
	}
}

func (log *CsiLogger) Flush() {
	glog.Flush()
}

func InitLogger() {
	_ = flag.Set("alsologtostderr", "false")
	_ = flag.Set("stderrthreshold", "INFO")
	_ = flag.Set("v", "4")
	_ = flag.Set("log_dir", "/var/log/")
	flag.Parse()
}
