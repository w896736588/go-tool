package gstool

import (
	"errors"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogPathFormat 日志路径格式类型
type LogPathFormat int

const (
	// LogPathFormatYearMonthDayLevel 格式: /var/log/test/2025/01/02debug.log
	LogPathFormatYearMonthDayLevel LogPathFormat = iota
	// LogPathFormatYearMonthLevelDay 格式: /var/log/test/2025/01/debug/02.log
	LogPathFormatYearMonthLevelDay
	// LogPathFormatYearMonthDay 格式: /var/log/test/2025/01/02.log
	LogPathFormatYearMonthDay
)

// GsSlog 基于zap的日志结构体
type GsSlog struct {
	logger                 *zap.Logger
	sugar                  *zap.SugaredLogger
	LogPath                string
	BusinessName           string
	BoolSimpleCallPosition bool
	logPathFormat          LogPathFormat
	currentDate            string
	atomLevel              zap.AtomicLevel
	dateLoggerMutex        sync.Mutex
	logFiles               []*os.File // 存储打开的文件句柄，用于在日期变化时关闭
	timer                  *time.Ticker
}

// NewSlog1 创建默认日志实例 /var/log/test/2025/01/02debug.log
func NewSlog1(logPath, businessName string) *GsSlog {
	return SlogCreate(logPath, businessName, true, LogPathFormatYearMonthDayLevel)
}

// NewSlog2 初始化默认日志实例 /var/log/test/2025/01/debug/02.log
func NewSlog2(logPath, businessName string) *GsSlog {
	return SlogCreate(logPath, businessName, true, LogPathFormatYearMonthLevelDay)
}

// NewSlog3 初始化默认日志实例 /var/log/test/2025/01/02.log
func NewSlog3(logPath, businessName string) *GsSlog {
	return SlogCreate(logPath, businessName, true, LogPathFormatYearMonthDay)
}

// SlogCreate 创建自定义日志实例
func SlogCreate(logPath, businessName string, boolSimpleCallPosition bool, logPathFormat LogPathFormat) *GsSlog {
	// 检测是否存在目录
	dirPath := logPath
	dirError := DirCreatePath(dirPath)
	if dirError != nil {
		panic(`创建目录失败 ` + dirError.Error())
	}
	currentDate := TimeNowUnixToString("Ymd")
	atomLevel := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	gsSlog := &GsSlog{
		LogPath:                logPath,
		BusinessName:           businessName,
		BoolSimpleCallPosition: boolSimpleCallPosition,
		logPathFormat:          logPathFormat,
		currentDate:            currentDate,
		atomLevel:              atomLevel,
		logFiles:               make([]*os.File, 0),
	}
	gsSlog.initLogger()
	return gsSlog
}

// closeLogFiles 关闭所有打开的日志文件
func (h *GsSlog) closeLogFiles() {
	for _, file := range h.logFiles {
		if file != nil {
			_ = file.Sync()  // 先同步数据到磁盘
			_ = file.Close() // 关闭文件
		}
	}
	h.logFiles = make([]*os.File, 0)
}

// initLogger 初始化日志记录器
func (h *GsSlog) initLogger() {
	h.closeLogFiles()

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      zapcore.OmitKey, // 移除日志级别显示
		NameKey:       "logger",
		CallerKey:     "caller",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		}, // 自定义时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			filename := filepath.Base(caller.File)
			enc.AppendString(filename + ":" + cast.ToString(caller.Line))
		},
	}
	if h.logPathFormat == LogPathFormatYearMonthDay {
		encoderConfig.LevelKey = "level"
	}
	var cores []zapcore.Core
	levels := []zapcore.Level{zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel, zapcore.ErrorLevel}
	for _, level := range levels {
		writer, file := h.getLogWriter(level)
		if file != nil {
			h.logFiles = append(h.logFiles, file)
		}
		levelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl == level
		})
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(writer),
			levelEnabler,
		)
		cores = append(cores, core)
	}
	core := zapcore.NewTee(cores...)
	h.logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	h.sugar = h.logger.Sugar()
}

// getLogWriter 获取日志写入器
func (h *GsSlog) getLogWriter(level zapcore.Level) (zapcore.WriteSyncer, *os.File) {
	currentDate := TimeNowUnixToString("Ymd")
	t, _ := TimeStringToUnix(currentDate, "Ymd")
	year := TimeUnixToString(t, "Y")
	month := TimeUnixToString(t, "m")
	day := TimeUnixToString(t, "d")

	var levelStr string
	switch level {
	case zapcore.InfoLevel:
		levelStr = "info"
	case zapcore.WarnLevel:
		levelStr = "warn"
	case zapcore.ErrorLevel:
		levelStr = "error"
	default:
		levelStr = "debug"
	}

	var logFilePath string
	switch h.logPathFormat {
	case LogPathFormatYearMonthDayLevel:
		// 格式: /var/log/test/2025/01/02debug.log
		datePath := filepath.Join(h.LogPath, h.BusinessName, year, month)
		_ = DirCreatePath(datePath)
		logFilePath = filepath.Join(datePath, day+levelStr+".log")

	case LogPathFormatYearMonthLevelDay:
		// 格式: /var/log/test/2025/01/debug/02.log
		datePath := filepath.Join(h.LogPath, h.BusinessName, year, month, levelStr)
		_ = DirCreatePath(datePath)
		logFilePath = filepath.Join(datePath, day+".log")

	case LogPathFormatYearMonthDay:
		// 格式: /var/log/test/2025/01/02.log
		datePath := filepath.Join(h.LogPath, h.BusinessName, year, month)
		_ = DirCreatePath(datePath)
		logFilePath = filepath.Join(datePath, day+".log")
	}

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return zapcore.AddSync(os.Stdout), nil
	}
	return zapcore.AddSync(logFile), logFile
}

// checkDateChange 检查日期是否变化，如果变化则更新日志文件
func (h *GsSlog) checkDateChange() {
	// 获取当前日期
	currentDate := TimeNowUnixToString("Ymd")

	// 如果日期没变，直接返回
	if currentDate == h.currentDate {
		return
	}
	h.dateLoggerMutex.Lock()
	defer h.dateLoggerMutex.Unlock()
	if currentDate == h.currentDate {
		return
	}
	h.currentDate = currentDate
	h.initLogger()
}

// Debugf 输出Debug级别日志
func (h *GsSlog) Debugf(msg string, args ...interface{}) {
	h.checkDateChange()
	h.sugar.Debugf("\n"+msg, args...)
}

// Infof 输出Info级别日志
func (h *GsSlog) Infof(msg string, args ...interface{}) {
	h.checkDateChange()
	h.sugar.Infof("\n"+msg, args...)
}

// Warnf 输出Warn级别日志
func (h *GsSlog) Warnf(msg string, args ...interface{}) {
	h.checkDateChange()
	h.sugar.Warnf("\n"+msg, args...)
}

// Errof 输出Error级别日志
func (h *GsSlog) Errof(msg string, args ...interface{}) {
	h.checkDateChange()
	h.sugar.Errorf("\n"+msg, args...)
}

// SetLogLevel 设置日志级别
func (h *GsSlog) SetLogLevel(level string) {
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn", "warning":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}
	h.atomLevel.SetLevel(zapLevel)
}

// SetLogPathFormat 设置日志路径格式
func (h *GsSlog) SetLogPathFormat(format LogPathFormat) {
	h.logPathFormat = format
	h.initLogger()
}

// Sync 同步缓冲区数据到磁盘
func (h *GsSlog) Sync() error {
	return h.logger.Sync()
}

// Close 关闭日志记录器并释放资源
func (h *GsSlog) Close() error {
	err := h.Sync()
	h.closeLogFiles()
	if h.timer != nil {
		h.timer.Stop()
	}
	return err
}

// CleanOldLogs 清理指定天数之前的旧日志
// keepDays: 保留最近多少天的日志（0表示不清理）
func (h *GsSlog) CleanOldLogs(keepDays int) error {
	if keepDays < 1 {
		return errors.New(`保留天数至少1天`)
	}
	go h.cleanLogs(keepDays)
	go func() {
		h.timer = time.NewTicker(1 * time.Hour) // 每小时触发一次
		for {
			select {
			case _ = <-h.timer.C:
				h.cleanLogs(keepDays)
			}
		}
	}()
	return nil
}

func (h *GsSlog) cleanLogs(keepDays int) {
	//这个日期及之前的全部清理
	cutoffTime := time.Now().AddDate(0, 0, -keepDays)
	cutoffDate := cutoffTime.Format("20060102")
	//检查的目录
	logRootPath := filepath.Join(h.LogPath, h.BusinessName)
	var err error
	//清理文件
	err = DirWalk(logRootPath, func(path string, info os.FileInfo, err error) {
		if info.IsDir() {
			return
		}
		//非日志文件不处理
		if !strings.HasSuffix(strings.ToLower(path), `.log`) {
			return
		}

		var year, month, day string
		filePath, _ := strings.CutPrefix(path, logRootPath)
		pathParams := strings.Split(filePath, string(filepath.Separator))
		switch h.logPathFormat {
		case LogPathFormatYearMonthDayLevel, LogPathFormatYearMonthDay: //2025/01/02.log: //var/log/test/2025/01/02debug.log
			if len(pathParams) != 4 {
				h.Errof(`按照当前模式获取的日志路径错误 %d %s`, h.logPathFormat, filePath)
				return
			}
			year = pathParams[1]
			month = pathParams[2]
			day = pathParams[3][0:2]
		case LogPathFormatYearMonthLevelDay: ///var/log/test/2025/01/debug/02.log
			if len(pathParams) != 5 {
				h.Errof(`按照当前模式获取的日志路径错误 %d %s`, h.logPathFormat, filePath)
				return
			}
			year = pathParams[1]
			month = pathParams[2]
			day = pathParams[4][0:2]
		}
		date := year + month + day
		if cutoffDate >= date {
			delErr := FileDelete(path)
			if delErr != nil {
				h.Errof(`清理日志 (%s) 失败 :%s`, path, delErr.Error())
			}
		}
	})
	if err != nil {
		h.Errof(`清理日志 (%s) 循环失败 :%s`, filepath.Join(h.LogPath, h.BusinessName), err.Error())
	}
	//清理空目录
	checkDelDirList := make([]string, 0)
	err = DirWalk(logRootPath, func(path string, info os.FileInfo, err error) {
		if !info.IsDir() {
			return
		}
		checkDelDirList = append(checkDelDirList, path)
	})
	if err != nil {
		h.Errof(`清理空目录 (%s) 失败 :%s`, filepath.Join(h.LogPath, h.BusinessName), err.Error())
		return
	}
	//循环删除空目录
	for i := len(checkDelDirList) - 1; i >= 0; i-- {
		checkDirPath := checkDelDirList[i]
		isEmpty, emptyErr := DirIsEmpty(checkDirPath)
		if emptyErr != nil {
			h.Errof(`判断目录 (%s) 是否为空失败 :%s`, checkDirPath, emptyErr.Error())
			continue
		}
		if isEmpty {
			removeErr := DirRemoveEmpty(checkDirPath)
			if removeErr != nil {
				h.Errof(`清理空目录 (%s) 失败 :%s`, checkDirPath, removeErr.Error())
			}
		}
	}
	return
}
