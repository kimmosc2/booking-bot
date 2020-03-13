package main

import (
	"booking-bot/config"
	"booking-bot/datasource"
	"booking-bot/flags"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)
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
		fmt.Println("BookingBot 200312.2103.Pro_beta\nAuthor:BuTn<kimmosc2@163.com>")
		os.Exit(0)
	}
	if flags.Help {
		flag.Usage()
		os.Exit(0)
	}
	users := datasource.NewConfiguration()
	config.LoadConfig(users)
	// 登录区
	loginChan := make(chan struct{}, 1)
	for _, u := range users.Users {
		go u.Login(loginChan)
	}
	for i := len(users.Users); i != 0; i-- {
		<-loginChan
	}
	fmt.Println("所有用户都已装载完毕")

	wgp := new(sync.WaitGroup)
	for _, u := range users.Users {
		wgp.Add(1)
		go u.FetchId(wgp)
	}
	wgp.Wait()

	fmt.Println("完成,抢购结果可扫描二维码进行查看,程序将在30秒后结束")
	time.Sleep(time.Second * 30)
}


