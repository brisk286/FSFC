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
	logger.InitLogger(config.GetConfig().Log.Path, config.GetConfig().Log.Level)

	newRouter := router.NewRouter()

	go server.MyServer.Start()

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
