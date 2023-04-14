package lib_tool

import (
	"fmt"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

type LibLog struct {
	Logger   *logrus.Logger
	LogPath  string
	FileName string
}

//初始化日志
func CreateLogger(logPath, fileName string) *LibLog {
	logger := logrus.New()
	libLogLogger := &LibLog{Logger: logger, LogPath: logPath, FileName: fileName}
	libLogLogger.InitDefault()
	return libLogLogger
}

//默认值初始化
func (h *LibLog) InitDefault() {
	//设置日志格式
	h.setTextFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	h.SetLogLevel(logrus.DebugLevel)
	h.SetRotateLogs(7, time.Hour*24*7, time.Hour*24)
}

//设置格式
func (h *LibLog) setTextFormatter(textFormat *logrus.TextFormatter) {
	h.Logger.SetFormatter(textFormat)
}

//设置日志等级
func (h *LibLog) SetLogLevel(logLevel logrus.Level) {
	h.Logger.SetLevel(logLevel)
}

//设置输出目录
func (h *LibLog) setOutputPath(logPath string) {
	src, err := setOutputFile(logPath)
	if err != nil {
		panic(fmt.Sprintf(`初始化日志失败 %s`, err.Error()))
	}
	//设置输出
	h.Logger.SetOutput(src)
}

//转到标准输出
func (h *LibLog) setOutputStdout() {
	h.Logger.SetOutput(os.Stdout)
}

//设置日志切割
func (h *LibLog) SetRotateLogs(maxCount int, maxAge, maxRotationTime time.Duration) {
	//检测是否存在目录失败
	dirPath := h.LogPath
	dirError := DirCreatePath(dirPath)
	if dirError != nil {
		panic(`创建目录失败 ` + dirError.Error())
	}
	fmt.Println(`设置文件 ` + dirPath + `/` + h.FileName + `%Y%m%d.log`)
	writer, err := rotatelogs.New(
		dirPath+`/`+h.FileName+`%Y%m%d.log`,
		//rotatelogs.WithLinkName(path),				//为最新的日志建立软连接(软链接对应的始终是最新日志)
		rotatelogs.WithRotationCount(maxCount),       //日志最多保留几个
		rotatelogs.WithMaxAge(maxAge),                //日志最多保留多久的
		rotatelogs.WithRotationTime(maxRotationTime), //日志多长时间创建一个
	)
	if err != nil {
		panic(fmt.Sprintf(`初始化日志切割错误 %#v`, err))
	}
	h.Logger.SetOutput(writer)
}

//设置输出到文件
func setOutputFile(logFilePath string) (*os.File, error) {
	now := time.Now()
	_, err := os.Stat(logFilePath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(logFilePath, 0777); err != nil {
			panic(fmt.Sprintf(`给与日志文件权限失败 %s`, err.Error()))
			return nil, err
		}
	}
	logFileName := now.Format("2006-01-02") + ".log"
	//日志文件
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			panic(fmt.Sprintf(`创建日志文件失败 %s`, err.Error()))
			return nil, err
		}
	}
	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(fmt.Sprintf(`打开日志文件失败 %s`, err.Error()))
		return nil, err
	}
	return src, nil
}

func (h *LibLog) Debugf(msg string, args ...interface{}) {
	if len(args) == 0 {
		h.Logger.Debug(msg)
	} else {
		h.Logger.Debugf(msg, args)
	}
}

func (h *LibLog) Infof(msg string, args ...interface{}) {
	if len(args) == 0 {
		h.Logger.Info(msg)
	} else {
		h.Logger.Infof(msg, args)
	}
}

func (h *LibLog) Warningf(msg string, args ...interface{}) {
	if len(args) == 0 {
		h.Logger.Warn(msg)
	} else {
		h.Logger.Warningf(msg, args)
	}
}

func (h *LibLog) Errorf(msg string, args ...interface{}) {
	if len(args) == 0 {
		h.Logger.Error(msg)
	} else {
		h.Logger.Errorf(msg, args)
	}
}
