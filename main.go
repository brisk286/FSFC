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

	newRouter := router.NewRouter()

	go router.MyServer.Start()

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
