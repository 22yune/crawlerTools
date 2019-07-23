package support

import (
	"net/http"
	"strings"
	"time"
	"net"
	"context"
	"errors"
	"fmt"
)

func Request(method string, url string, body string, cookies map[string]string, heads map[string]string, timeOut time.Duration, retry int) (resp *http.Response, err error) {

	defer func() {
		if p := recover(); p!= nil{
			err = errors.New(fmt.Sprint(p))
		}
	}()

	if timeOut == 0 {
		timeOut = 30 * time.Second
	}
	var myDial = &net.Dialer{
		Timeout:   timeOut,
		KeepAlive: timeOut,
		DualStack: true,
	}

	// 自定义DialContext
	var myDialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		network = "tcp4" //仅使用ipv4
		//network = "tcp6" //仅使用ipv6
		return myDial.DialContext(ctx, network, addr)
	}

	var client = &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           myDialContext,
			MaxIdleConns:          20,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, _ := http.NewRequest(strings.ToUpper(method), url, strings.NewReader(body))

	for k, v := range cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	heads["Accept"] = "*/*"
	heads["User-Agent"] = "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:66.0) Gecko/20100101 Firefox/66.0"

	if _, ok := heads["Content-Type"]; !ok {
		heads["Content-Type"] = "application/x-www-form-urlencoded; charset=UTF-8"
	}
	for k, v := range heads {
		req.Header.Add(k, v)
	}

	resp, err = client.Do(req)
	if err != nil && retry > 0 {
		return Request(method, url, body, cookies, heads, timeOut*2, retry-1)
	}
	return resp, err
}
