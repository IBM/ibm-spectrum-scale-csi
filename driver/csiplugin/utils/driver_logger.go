package utils

import (
	"flag"
	"github.com/golang/glog"
)

type CsiLogger struct{}

func (log *CsiLogger) Infof(format string, args ...interface{}) {
	glog.Infof(format, args)
}

func (log *CsiLogger) Debugf(format string, args ...interface{}) {
	glog.V(4).Infof(format, args)
}

func (log *CsiLogger) DebugPlus(format string, args ...interface{}) {
	glog.V(6).Infof(format, args)
}

func (log *CsiLogger) Errorf(format string, args ...interface{}) {
	glog.Errorf(format, args)
}

func (log *CsiLogger) Fatalf(format string, args ...interface{}) {
	glog.Fatalf(format, args)
}

func (log *CsiLogger) Flush() {
	glog.Flush()
}

func InitLogger() {
	_ = flag.Set("alsologtostderr", "true")
	_ = flag.Set("stderrthreshold", "ERROR")
	_ = flag.Set("log_dir", "/var/log/")
	flag.Parse()
}
