package utils

import (
	"io"
	"log"
	"os"
)

var Logger *log.Logger
var logFile *os.File

func init() {
	// 默认使用空输出，避免强制创建日志文件
	Logger = log.New(io.Discard, "", log.Ltime)
}

// InitLogger 初始化日志记录器
// enable: 是否启用日志记录
func InitLogger(enable bool) error {
	if !enable {
		Logger = log.New(io.Discard, "", log.Ltime)
		return nil
	}

	var err error
	logFile, err = os.OpenFile("file_finder.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	Logger = log.New(logFile, "", log.Ltime)
	return nil
}

// CloseLogger 关闭日志文件
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
