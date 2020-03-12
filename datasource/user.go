package datasource

import (
	"net/http"
)

// 用户信息
type User struct {
	// 账号
	Account string `json:"account"`
	// 密码
	Password string `json:"password"`

	Client *http.Client
	// 登录状态
	IsLogin bool
	// 市
	Country string `json:"country"`
	// 详细地址
	Address string `json:"address"`
	// 省
	City string `json:"city"`
	// 姓名
	Name string `json:"name"`
	// userId,购买时需要的
	Id string `json:"id"`
	// Token
	Token string `json:"token"`
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
}
