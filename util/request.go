package util

import (
	"net/http"
)





// SetAutoRequestHeader set some common header options
func SetAutoRequestHeader(req *http.Request, token string) {
	req.Header.Set("token", token)
	req.Header.Set("Accept", `application/json, text/plain, */*`)
	req.Header.Set("Accept-Encoding", `gzip, deflate`)
	req.Header.Set("Accept-Language", `zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6`)
	req.Header.Set("Connection", `keep-alive`)
	req.Header.Set("Content-Type", `application/json;charset=UTF-8`)
	// req.Header.Set("Referer", `http://reserveyd.inca.com.cn/index?companyid=`)
	req.Header.Set("User-Agent", `Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1`)
}
