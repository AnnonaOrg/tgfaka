package log

import (
	"fmt"

	"github.com/umfaka/tgfaka/internal/exts/config"

	// "github.com/AnnonaOrg/osenv"
	"github.com/sirupsen/logrus"
)

func Init() {
	logLevelStr := config.GetSiteConfig().LogLevel // osenv.GetLogLevel()
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		fmt.Printf("日志级别(%s)解析失败❌: %v\n", logLevelStr, err)
		return
	}
	logrus.SetLevel(logLevel)
	// 日志输出文件位置
	// logrus.SetReportCaller(true)
	// Set the formatter to include file and line information
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 MST",
		// CallerPrettyfier: func(f *runtime.Frame) (string, string) {
		// 	return "", fmt.Sprintf("%s:%d", f.File, f.Line)
		// },
	})

	fmt.Println("日志配置完成✅")
	return
}

// Trace logs a message at level Trace on the standard logger.
func Trace(args ...interface{}) {
	logrus.Trace(args...)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	logrus.Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	logrus.Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	logrus.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(args ...interface{}) {
	logrus.Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	logrus.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	logrus.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

// Tracef logs a message at level Trace on the standard logger.
func Tracef(format string, args ...interface{}) {
	logrus.Tracef(format, args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(format string, args ...interface{}) {
	logrus.Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func Warningf(format string, args ...interface{}) {
	logrus.Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...interface{}) {
	logrus.Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

// Traceln logs a message at level Trace on the standard logger.
func Traceln(args ...interface{}) {
	logrus.Traceln(args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...interface{}) {
	logrus.Debugln(args...)
}

// Println logs a message at level Info on the standard logger.
func Println(args ...interface{}) {
	logrus.Println(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...interface{}) {
	logrus.Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...interface{}) {
	logrus.Warnln(args...)
}

// Warningln logs a message at level Warn on the standard logger.
func Warningln(args ...interface{}) {
	logrus.Warningln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...interface{}) {
	logrus.Errorln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(args ...interface{}) {
	logrus.Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatalln(args ...interface{}) {
	logrus.Fatalln(args...)
}
