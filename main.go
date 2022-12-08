package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unsafe"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Ahunet struct {
	Username string
	Password string
	base     string
	client   *http.Client
}

// return the ahunet struct
func NewAhuNet(username, password string) Ahunet {
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Proxy: nil,
		},
	}
	return Ahunet{
		Username: username,
		Password: password,
		base:     "http://172.16.253.3",
		client:   client,
	}
}

// get the ipv4 address info
func (ahu *Ahunet) GetIpv4Info() string {
	resp, err := ahu.client.Get(ahu.base + "/drcom/chkstatus?callback=dr1002&v=123")
	if err != nil {
		logrus.Errorln("get info error:", err)
		return ""
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorln("read stream body err:", err)
		return ""
	}
	s := dealJsonP(*(*string)(unsafe.Pointer(&bs)))
	ipv4 := gjson.Get(s, "v46ip").String()
	return ipv4

}

// authentication
func (ahu *Ahunet) Auth(ipv4 string) {
	q := url.Values{}
	q.Add("c", "Portal")
	q.Add("a", "login")
	q.Add("callback", "dr1003")
	q.Add("login_method", "1")
	q.Add("user_account", ahu.Username)
	q.Add("user_password", ahu.Password)
	q.Add("wlan_user_ip", ipv4)
	q.Add("wlan_user_ipv6", "")
	q.Add("wlan_user_mac", "000000000000")
	q.Add("wlan_ac_ip", "")
	resp, err := ahu.client.Get(ahu.base + ":801/eportal/?" + q.Encode())
	if err != nil {
		logrus.Errorln("auth error:", err)
		return
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorln("read stream body err:", err)
	}
	s := dealJsonP(*(*string)(unsafe.Pointer(&bs)))
	result := gjson.Parse(s)
	code := result.Get("ret_code").Int()
	switch code {
	case 1:
		color.Green("认证成功! IP: %s\n", ipv4)
	case 2:
		color.Yellow("已经在线! IP: %s\n", ipv4)
	default:
		fmt.Println(result.Get("msg").String())
	}
}

// trim left bracket and right bracket
func dealJsonP(origin string) string {
	left := strings.Index(origin, "(")
	right := strings.LastIndex(origin, ")")
	return origin[left+1 : right]
}

var username = flag.String("u", "", "username")
var password = flag.String("p", "", "password")

func main() {
	flag.Parse()
	ahu := NewAhuNet(*username, *password)
	ipv4 := ahu.GetIpv4Info()
	ahu.Auth(ipv4)
}
