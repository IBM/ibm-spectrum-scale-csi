package utils

type Logger interface {
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	DebugPlus(format string, args ...interface{})
	Flush()
}
