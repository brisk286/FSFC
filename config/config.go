package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type TomlConfig struct {
	AppName    string
	Log        LogConfig
	StaticPath PathConfig
}

// 基础设置
type Setting struct {
	ScanGap string
}

// 日志保存地址
type LogConfig struct {
	Path  string
	Level string
}

// 相关地址信息，例如静态文件地址
type PathConfig struct {
	FilePath string
}

var c TomlConfig

func init() {
	// 设置文件名
	viper.SetConfigName("config")
	// 设置文件类型
	viper.SetConfigType("toml")
	// 设置文件路径，在工作目录中查找配置
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		return
	}
	fmt.Println(c)
}

func GetConfig() TomlConfig {
	return c
}
