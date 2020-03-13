package config

import (
	"booking-bot/datasource"
	"booking-bot/flags"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// 配置文件
type UserConf struct {
	// 手机号
	Account string `json:"account"`
	// 密码
	Password string `json:"password"`
	// 药店id
	PlacePointId string `json:"placePointId"`
	// 间隔时间
	Sleep int `json:"sleep"`
	// 最大重试次数
	Max int `json:"max"`
	// 预约个数
	Num int `json:"num"`
	// 地区id
	CompanyId string `json:"companyId"`
	// Goroutine数量
	Grade int `json:"grade"`

}

// 加载配置文件
func LoadConfig(users *datasource.Configuration) {
	open, err := os.Open(flags.ConfigPath)
	if err != nil {
		log.Fatal("open file error:", err.Error())
	}
	b, _ := ioutil.ReadAll(open)
	err = json.Unmarshal(b, users)
	if err != nil {
		log.Fatal("parse config error:", err.Error())
	}

}
