package stdlib

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func getEncoder(args ...string) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(time.Now().UTC().Format("2006-01-02T15:04:05.999999")) // this is the format of the time added to the log
		for _, arg := range args {
			enc.AppendString(arg)
		}
		//You can add more strings to log by using enc.AppendString("whatever you want")
	})
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.CallerKey = ""
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func initLogger(args ...string) { // for logging to the console
	encoder := getEncoder(args...)
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.DebugLevel)
	logg := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logg)
}

func initLoggerWithFile(args ...string) {
	createDirectoryIfNotExists()
	writerSync := getLogWriter()
	encoder := getEncoder(args...)
	core := zapcore.NewCore(encoder, writerSync, zap.DebugLevel)
	logg := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logg)
}

func createDirectoryIfNotExists() {
	path, _ := os.Getwd()
	if _, err := os.Stat(path + "/logs"); os.IsNotExist(err) {
		os.Mkdir(path+"/logs", os.ModePerm)
	}
}
func getLogWriter() zapcore.WriteSyncer {
	createDirectoryIfNotExists()
	path, _ := os.Getwd()
	return zapcore.AddSync(&lumberjack.Logger{ //found this cool package that manages the log files
		Filename:   path + "/logs/log.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	})
}
func ConsoleLogger(args ...string) *zap.SugaredLogger {
	initLogger(args...)
	return zap.S()
}

func FileLogger(args ...string) *zap.SugaredLogger {
	initLoggerWithFile(args...)
	return zap.S()
}
