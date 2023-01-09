package utils

import (
	"context"
	"flag"
	"k8s.io/klog/v2"
	"os"
    	"path/filepath"
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
	klog.Infof("logValue: %s", logValue)

	dirPath := "/host/var/log/" + path + "/"
	if !Exists(ctx, dirPath) {
		err := MkDir(ctx, dirPath)
		if err != nil {
			klog.Errorf("Failed to create log directory")
		}
	}
	path := filepath.Join(dirPath,"ibm-spectrum-scale-csi.log")

	var fs flag.FlagSet
    	klog.InitFlags(&fs)
    	fs.Set("logtostderr", "false")
	fs.Set("stderrthreshold", logValue)
    	fs.Set("log_file", path)
	if level == DEBUG.String() {
        	fs.Set("v", "4")
        } else if level == TRACE.String() {
                fs.Set("v", "6")
        } else {
                fs.Set("v", "1")
        }

}
