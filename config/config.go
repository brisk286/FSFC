package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type TomlConfig struct {
	AppName    string
	Set        Setting
	Web        WebConfig
	Log        LogConfig
	StaticPath PathConfig
}

// Setting 基础设置
type Setting struct {
	ScanGap    int
	LocalPath  string
	RemotePath string
}

// WebConfig 网络设置
type WebConfig struct {
	LocalIp    string
	LocalPort  string
	RemoteIp   string
	RemotePort string
}

// LogConfig 日志保存地址
type LogConfig struct {
	Path  string
	Level string
}

// PathConfig 相关地址信息，例如静态文件地址
type PathConfig struct {
	FilePath string
}

var c TomlConfig

func init() {
	// 设置文件名
	//viper.SetConfigName("config")
	viper.SetConfigName("config_test")
	// 设置文件类型
	viper.SetConfigType("toml")
	// 设置文件路径，在工作目录中查找配置
	viper.AddConfigPath("../")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		return
	}
	//fmt.Println(c)
}

func GetConfig() TomlConfig {
	return c
}
