package main

import (
	config2 "booking-bot/config"
	"booking-bot/flags"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
	"time"
	// "github.com/peterh/liner"
)

// TODO HeaderFactory

// 登录表单
type loginForm struct {
	UserForm  `json:"data"`
	CompanyId string `json:"companyid"`
}

// 手机号，密码
type UserForm struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// 登录返回信息
type Result struct {
	Data   User   `json:"data"`
	Msg    string `json:"msg"`
	Status int    `json:"status"`
}

// 用户信息(登录后返回的)
type User struct {
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
}

// 每天id结构体解码
type TodayList struct {
	Data TodayData `json:"data"`
}

type TodayData struct {
	TodayGoodsDtos []TodayGoodsDtos `json:"todayGoodsDtos"`
}

type TodayGoodsDtos struct {
	// 药店名
	PlacepointName string `json:"placepointName"`
	// 口罩id
	Id int `json:"id"`
}

// 抢购返回json
type Response struct {
	Data   string `json:"data"`
	Msg    string `json:"msg"`
	Status int    `json:"status"`
}

var (
	client = new(http.Client)
	// 登录凭证
	token      string
	loginData  = new(loginForm)
	config     = new(config2.Config)
	isLogin    = false
	userInfo   = new(User)
	targetInfo = new(TodayList)
)

// cookie 初始化
func cookieInit() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal("初始化cookie失败:", err.Error())
	}
	client.Jar = jar
}

// 表单初始化
func formInit() {
	loginData.CompanyId = config.CompanyId
}

func parseConfig() {

	open, err := os.Open(flags.ConfigPath)
	if err != nil {
		log.Fatal("open file error:", err.Error())
	}
	b, _ := ioutil.ReadAll(open)
	err = json.Unmarshal(b, &config)
	if err != nil {
		log.Fatal("parse config error:", err.Error())
	}
	if config.Num > 5 {
		log.Fatal("最大抢购数量为5")
	}
	if config.CompanyId == "" {
		log.Fatal("请填写地区id")
	}
}

// 获取当日口罩id, data是药店id, id是购买者id,token是登录token
func fetchId(data, id, token, companyId string) {
	body := `{"companyId":"` + companyId + `","data":"` + data + `","loginid":"` + id + `"}`
	req, err := http.NewRequest(http.MethodPost, "http://reservepc.inca.com.cn:9345/booking/todayList", strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", `application/json, text/plain, */*`)
	req.Header.Set("Accept-Encoding", `gzip, deflate`)
	req.Header.Set("Accept-Language", `zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6`)
	req.Header.Set("Connection", `keep-alive`)
	req.Header.Set("Content-Type", `application/json;charset=UTF-8`)
	req.Header.Set("Referer", `http://reserveyd.inca.com.cn/index?companyid=`+companyId)
	req.Header.Set("User-Agent", `Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1`)
	req.Header.Set("token", token)

	for {
		fmt.Println("正在获取药店信息...")
		do, err := client.Do(req)
		if err == nil {
			b, _ := ioutil.ReadAll(do.Body)
			err := json.Unmarshal(b, &targetInfo)
			if err != nil {
				log.Fatal("unmarshal failed:", err.Error())
			}
			if len(targetInfo.Data.TodayGoodsDtos) > 0 && targetInfo.Data.TodayGoodsDtos[0].Id != 0 {
				break
			}
		}
	}
}

func init() {
	flag.Parse()
	file := "./" + "message" + ".txt"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // 将文件设置为log输出的文件
	log.SetPrefix("[booking-bot]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
}

func main() {
	if flags.Version {
		fmt.Println("BookingBot_200310_1720_alpha")
		os.Exit(0)
	}
	if flags.Help {
		flag.Usage()
		os.Exit(0)
	}

	cookieInit()
	formInit()
	parseConfig()
	fmt.Println("指定的配置文件:", flags.ConfigPath)
	targetInfo.Data.TodayGoodsDtos = make([]TodayGoodsDtos, 1)
	var user UserForm
	user.Account = config.Account
	user.Password = config.Password
	loginData.UserForm = user
	marshal, _ := json.Marshal(loginData)
	// 登录
	for !isLogin {
		Login(string(marshal))
	}
	fetchId(config.PlacePointId, userInfo.Id, token, config.CompanyId)
	fmt.Println("本次预约:", targetInfo.Data.TodayGoodsDtos[0].PlacepointName)
	// 买
	realBuy(strconv.Itoa(config.Num), config.PlacePointId, strconv.Itoa(targetInfo.Data.TodayGoodsDtos[0].Id), config.CompanyId, config.Sleep, config.Max)

}

func realBuy(num string, placePointId string, saleGoodsId, companyId string, sleep, max int) {
	for i := 1; i <= max; i++ {
		fmt.Printf("第%d次尝试抢购:", i)
		// 构造购买请求
		buy := Buy(num, placePointId, saleGoodsId, userInfo.Id, companyId)

		// 发起购买
		resp, err := client.Do(buy)
		if err != nil {
			panic(err)
		}
		b, _ := ioutil.ReadAll(resp.Body)
		response := new(Response)
		json.Unmarshal(b, &response)
		fmt.Println(response.Msg) // 抢购结果
		if response.Status == 200 {
			break
		}
		resp.Body.Close()
		time.Sleep(time.Millisecond * time.Duration(sleep))
	}
	fmt.Println("完成,抢购结果可扫描二维码进行查看,程序将在30秒后结束")
	time.Sleep(time.Second * 30)
}

// 登录
func Login(body string) {
	url := `http://reservepc.inca.com.cn:9345/login/auth`
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		log.Fatal("登录请求初始化失败:", err.Error())
	}
	request.Header.Set("Connection", `keep-alive`)
	request.Header.Set("Accept", `application/json, text/plain, */*`)
	request.Header.Set("User-Agent", `Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1`)
	request.Header.Set("Content-Type", `application/json;charset=UTF-8`)
	request.Header.Set("Accept-Encoding", `gzip, deflate`)
	request.Header.Set("Accept-Language", `zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6`)
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal("在执行登录时出现了一个错误:", err.Error())
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)

	result := new(Result)
	json.Unmarshal(b, &result)

	if result.Status != 200 {
		fmt.Println("登陆失败:", result.Msg)
		isLogin = false
	} else {
		isLogin = true
		fmt.Printf("您好:%s\n", result.Data.Name)
		userInfo = &result.Data

		token = resp.Header.Get("Token")
	}

}

// 购买
func Buy(num, placePointId, saleGoodsId, userId, companyId string) *http.Request {
	body := `{"companyId":"` + companyId + `","data":{"address":"","deliverState":"0","goodsQty":` + num + `,"zdGoodsName":"一次性医用口罩","placepointId":"` + placePointId + `","saleGoodsId":` + saleGoodsId + `,"userId":"` + userId + `"}}`
	request, err := http.NewRequest(http.MethodPost, "http://reservepc.inca.com.cn:9345/booking/save", strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	request.Header.Set("token", token)
	request.Header.Set("Connection", `keep-alive`)
	request.Header.Set("Accept", `application/json, text/plain, */*`)
	request.Header.Set("User-Agent", `Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1`)
	request.Header.Set("Content-Type", `application/json;charset=UTF-8`)
	request.Header.Set("Origin", `http://reserveyd.inca.com.cn`)
	request.Header.Set("Referer", `http://reserveyd.inca.com.cn/index?companyid=0ac83ca4260a4549acfb383a14665783`)
	request.Header.Set("Accept-Encoding", `gzip, deflate`)
	request.Header.Set("Accept-Language", `zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6`)

	return request
}
