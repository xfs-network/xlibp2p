package log

import "github.com/sirupsen/logrus"

type Logger interface {
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Warningf(string, ...interface{})
}

func DefaultLogger() Logger {
	logrus.SetLevel(logrus.DebugLevel)
	return logrus.StandardLogger()
}
