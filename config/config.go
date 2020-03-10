package config

// 配置文件
type Config struct {
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
}

