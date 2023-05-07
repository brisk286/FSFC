package main

import (
	"fsfc/config"
	"fsfc/logger"
	"fsfc/router"
	"net/http"
	"time"
)

func main() {
	logger.InitLogger(config.Config.Log.Path, config.Config.Log.Level)

	// 启动一个协程检测文件修改情况
	go router.MyServer.Start()

	// 启动server路由
	newRouter := router.NewRouter()
	s := &http.Server{
		Addr:           ":8888",
		Handler:        newRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := s.ListenAndServe()
	if nil != err {
		logger.Logger.Error("server error", logger.Any("serverError", err))
	}
}
