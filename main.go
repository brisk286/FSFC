package main

import (
	"fsfc/config"
	"fsfc/logger"
	"fsfc/router"
	"fsfc/server"
	"net/http"
	"time"
)

func main() {
	//配置logger
	logger.InitLogger(config.GetConfig().Log.Path, config.GetConfig().Log.Level)

	logger.Logger.Info("config", logger.Any("config", config.GetConfig()))

	//Logger
	logger.Logger.Info("start server", logger.String("start", "start web sever..."))

	//设置路由
	newRouter := router.NewRouter()

	//在本地开一个端口  接收信息
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

	//开启后端
	go server.MyServer.Start()
}
