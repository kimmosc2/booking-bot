package datasource

type Response struct {
	Data   interface{} `json:"data"`
	Msg    string      `json:"msg"`
	Status int         `json:"status"`
}

// 每天id结构体解码
type TodayList struct {
	Data TodayData `json:"data"`
}
// 每天id结构体解码
type TodayData struct {
	TodayGoodsDtos []TodayGoodsDtos `json:"todayGoodsDtos"`
}

type TodayGoodsDtos struct {
	// 预约物品名称
	ZdGoodsName string `json:"zdGoodsName"`
	// 药店名
	PlacepointName string `json:"placepointName"`
	// 口罩id
	Id int `json:"id"`
}
