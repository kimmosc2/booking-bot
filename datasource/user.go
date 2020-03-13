package datasource

import (
	"booking-bot/util"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"sync"
	"time"
)

func NewConfiguration() *Configuration {
	c := new(Configuration)
	c.Users = make([]*User, 1)
	return c
}

type Configuration struct {
	Users []*User `json:"users"`
}

// 用户信息
type User struct {
	// 账号
	Account string `json:"account"`
	// 密码
	Password string `json:"password"`
	// httpClient
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
	// 药店名称
	PlacepointName string `json:"placepointName"`
	// 间隔时间
	Sleep int `json:"sleep"`
	// 最大重试次数
	Max int `json:"max"`
	// 预约个数
	Num int `json:"num"`
	// 地区id
	CompanyId string `json:"companyId"`
	// 口罩id,每天变化,提前十分钟获取
	SaleGoodsId int `json:"saleGoodsId"`
	// 口罩名称
	ZgGoodsName string `json:"zdGoodsName"`
	// Goroutine数量
	Grade int `json:"grade"`
	// 登录次数
	loginCount int
	// 预约是否成功
	BookingSuccess bool `json:"success"`
	// 有id
	HasId bool `json:"hasId"`
}

type LoginForm struct {
	User      `json:"data"`
	CompanyId string `json:"companyid"`
}

// 内嵌循环,登录不成功就循环登录
func (u *User) Login(ch chan struct{}) {
	for !u.IsLogin {
		u.loginCount++
		url := `http://reservepc.inca.com.cn:9345/login/auth`
		form := new(LoginForm)
		bindForm(form, u)
		user, err := json.Marshal(form)
		if err != nil {
			log.Fatal("json marshal error:", err)
		}
		request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(user))
		if err != nil {
			log.Fatal("登录请求初始化失败:", err.Error())
		}
		util.SetAutoRequestHeader(request, "")
		if u.Client == nil {
			u.Client = new(http.Client)
			jar, _ := cookiejar.New(nil)
			u.Client.Jar = jar
		}

		resp, err := u.Client.Do(request)
		if err != nil {
			log.Fatal("在执行登录时出现了一个错误:", err.Error())
		}
		b, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		r := new(Response)
		r.Data = u
		err = json.Unmarshal(b, &r)
		if err != nil {
			log.Fatal("json unmarshal error:", err)
		}

		if r.Status != 200 {
			if u.loginCount >= 5 {
				fmt.Printf("%v已连续登录失败5次,已取消登录\n", u.Account)
				ch <- struct{}{}
				break
			}
			fmt.Printf("%v登陆失败:%v,正在重试\n", u.Account, r.Msg)
			u.IsLogin = false
		} else {
			u.IsLogin = true
			fmt.Printf("%v登录成功,姓名:%s\n", u.Account, u.Name)
			u.Token = resp.Header.Get("Token")
			ch <- struct{}{}
			break
		}
		time.Sleep(time.Second)
	}
}

func (u *User) FetchId(wgp *sync.WaitGroup) {
	if !u.IsLogin {
		return
	}
	body := `{"companyId":"` + u.CompanyId + `","data":"` + u.PlacePointId + `","loginid":"` + u.Id + `"}`
	req, err := http.NewRequest(http.MethodPost, "http://reservepc.inca.com.cn:9345/booking/todayList", strings.NewReader(body))
	if err != nil {
		log.Fatal("build request error:", err.Error())
	}
	util.SetAutoRequestHeader(req, u.Token)

	for {
		fmt.Printf("%v正在获取药店信息\n", u.Name)
		do, err := u.Client.Do(req)
		if err == nil {
			b, _ := ioutil.ReadAll(do.Body)
			t := new(TodayList)
			t.Data.TodayGoodsDtos = make([]TodayGoodsDtos, 1)
			err := json.Unmarshal(b, &t)
			if err != nil {
				log.Println("unmarshal saleid error:", err.Error())
				return
			}
			if len(t.Data.TodayGoodsDtos) > 0 && t.Data.TodayGoodsDtos[0].Id != 0 {
				u.ZgGoodsName = t.Data.TodayGoodsDtos[0].ZdGoodsName
				u.PlacepointName = t.Data.TodayGoodsDtos[0].PlacepointName
				u.SaleGoodsId = t.Data.TodayGoodsDtos[0].Id
				u.Buy()
				wgp.Done()
				break
			}
		}
		time.Sleep(time.Second)
	}
}

func (u *User) Buy() {
	if !u.IsLogin {
		return
	}
	count := 0
	wg := new(sync.WaitGroup)
	for i := 0; i < u.Grade; i++ {
		wg.Add(1)
		go func(worker int, c int, user *User, w *sync.WaitGroup) {
			for ; c <= user.Max; c++ {
				if user.BookingSuccess {
					w.Done()
					break
				}
				// 构造购买请求
				r := buildBookingRequest(strconv.Itoa(user.Num), user.PlacePointId, strconv.Itoa(user.SaleGoodsId), user.Id, user.CompanyId, user.Token, user.ZgGoodsName)
				realBuy(user, r)
				time.Sleep(time.Millisecond * time.Duration(user.Sleep))
			}
		}(i, count, u, wg)
	}
	wg.Wait()
}

func realBuy(u *User, req *http.Request) {
	resp, err := u.Client.Do(req)
	if err != nil {
		log.Println("request error:", err.Error())
		return
	}
	b, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	response := new(Response)
	err = json.Unmarshal(b, &response)
	if err != nil {
		log.Println("unmarshal error:", err.Error())
		return
	}

	if response.Status == 200 {
		fmt.Printf("%v抢购成功\n", u.Name)
		log.Printf("%v抢购成功\n", u.Name)
		u.BookingSuccess = true
		return
	}
	fmt.Printf("【%s】 | 【%s】 | 【%s】 | 【%s】  \n", u.Name, u.PlacepointName, u.ZgGoodsName, response.Msg) // 抢购结果
}

func buildBookingRequest(num, placePointId, saleGoodsId, userId, companyId, token, zgGoodsName string) *http.Request {
	body := `{"companyId":"` + companyId + `","data":{"address":"","deliverState":"0","goodsQty":` + num +
		`,"zdGoodsName":"` + zgGoodsName + `","placepointId":"` + placePointId + `","saleGoodsId":` + saleGoodsId +
		`,"userId":"` + userId + `"}}`
	request, err := http.NewRequest(http.MethodPost, "http://reservepc.inca.com.cn:9345/booking/save", strings.NewReader(body))
	if err != nil {
		log.Fatal("build Buy request error:", err.Error())
	}
	util.SetAutoRequestHeader(request, token)
	return request
}

func bindForm(form *LoginForm, u *User) {
	form.CompanyId = u.CompanyId
	form.Account = u.Account
	form.Password = u.Password
}
