package log //nolint:typecheck

import (
	"path/filepath"
	"time"

	"github.com/adamweixuan/getty"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	runLogger   getty.Logger
	traceLogger getty.Logger
	errorLogger getty.Logger
)

const (
	logDir   = "./output/log/"
	run      = "run.log"
	trace    = "trace.log"
	errorLog = "error.log"

	logTmFmtWithMS = "2006-01-02 15:04:05.000"
)

// debug 模式下会将日志输出到控制台和文件，其他模式只输出到文件
func getLogWriter(path string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func getEncoder() zapcore.Encoder {
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + t.Format(logTmFmtWithMS) + "]")
	}
	customLevelEncoder := func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + level.CapitalString() + "]")
	}

	customCallerEncoder := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + caller.TrimmedPath() + "]")
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.EncodeLevel = customLevelEncoder
	encoderConfig.EncodeCaller = customCallerEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func Init(level string, logPath string) {
	if len(logPath) == 0 {
		logPath = logDir
	}

	runLogger = initSugaredLogger(filepath.Join(logPath, run), true, level)
	errorLogger = initSugaredLogger(filepath.Join(logPath, errorLog), true, level)
	traceLogger = initSugaredLogger(filepath.Join(logPath, trace), false, level)
	// 框架的日志也输入到 run.log 中
	getty.SetLogger(runLogger)
}

func initLogger(path string, caller bool, level string) *zap.Logger {
	zapLevel := zapcore.InfoLevel
	// 忽略错误，如果传入的字符串有误默认 info 级别
	_ = zapLevel.Set(level)

	encoder := getEncoder()
	writeSyncer := getLogWriter(path)
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)
	if caller {
		return zap.New(core, zap.AddCaller())
	}
	return zap.New(core)
}

func initSugaredLogger(path string, caller bool, level string) *zap.SugaredLogger {
	logger := initLogger(path, caller, level)
	return logger.Sugar()
}

func Run() getty.Logger {
	return runLogger
}

func Trace() getty.Logger {
	return traceLogger
}

func Error() getty.Logger {
	return errorLogger
}
