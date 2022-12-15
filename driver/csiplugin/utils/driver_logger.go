package utils

import (
	"context"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"strings"
)

type CsiLogger struct{}

var logger *CsiLogger
var depth int = 1

func (log *CsiLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	arg := setFormat(ctx, format, args)
	glog.InfoDepth(depth, arg)
}

func (log *CsiLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	arg := setFormat(ctx, format, args)
	if glog.V(4) {
		glog.InfoDepth(depth, arg)
	}
}

func (log *CsiLogger) DebugPlus(ctx context.Context, format string, args ...interface{}) {
	arg := setFormat(ctx, format, args)
	if glog.V(6) {
		glog.InfoDepth(depth, arg)
	}
}

func (log *CsiLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	arg := setFormat(ctx, format, args)
	glog.ErrorDepth(depth, arg)
}

func (log *CsiLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	arg := setFormat(ctx, format, args)
	glog.FatalDepth(depth, arg)
}

func (log *CsiLogger) Flush() {
	glog.Flush()
}

func setFormat(ctx context.Context, format string, args ...interface{}) string {
	var arg []string
	var logFormat string
	for index, _ := range args {
		val := fmt.Sprintf("%v", args[index])
		arg = append(arg, val)
	}
	argsValue := strings.Join(arg, ",")
	if ctx != nil {
		loggerId := GetLoggerId(ctx)
		logFormat = fmt.Sprintf("[%s] %s", loggerId, format)
		return fmt.Sprintf(logFormat, argsValue)
	} else {
		return fmt.Sprintf(format, argsValue)
	}
}

func InitLogger() {
	_ = flag.Set("alsologtostderr", "false")
	_ = flag.Set("stderrthreshold", "INFO")
	_ = flag.Set("v", "4")
	_ = flag.Set("log_dir", "/var/log/")
	flag.Parse()
}
