package utils

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	logFile, err := os.OpenFile("file_finder.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}
	Logger = log.New(logFile, "", log.Ltime)
}
