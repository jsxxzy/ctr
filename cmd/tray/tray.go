package main

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/jsxxzy/ctr"
)

func main() {
	var App = newFreeApp()
	App.Loop()
}

// 获取可用端口
func getAvailablePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil

}

type app struct {
	Port         int         // 监听端口
	CtrServer    *ctr.Server // http-server
	serverStatus bool        // 服务状态
}

func newFreeApp() *app {
	var p, _ = getAvailablePort()
	return &app{
		Port:      p,
		CtrServer: ctr.New(p),
	}
}

func newApp(p int) *app {
	return &app{
		Port:      p,
		CtrServer: ctr.New(p),
	}
}

func (a *app) Init() {
	go func() {
		fmt.Println("start http server")
		a.Start()
	}()
	a.serverStatus = true
	systray.SetTitle("远程管理")
	ipButton := systray.AddMenuItem(a.GetIP(), "单击复制到剪贴板")
	systray.AddSeparator()
	var c = "检测服务"
	checkButton := systray.AddMenuItem(c, "检测服务是否正常")
	actionButton := systray.AddMenuItem("退出", "退出")
	go func() {
		for {
			select {
			case <-ipButton.ClickedCh:
				a.CopyIpv4()
			case <-checkButton.ClickedCh:
				var f = a.Ping()
				var s = "成功"
				if !f {
					s = "失败"
				}
				a.serverStatus = f
				var output = fmt.Sprintf("检测服务(上次%s)", s)
				checkButton.SetTitle(output)
				ipButton.SetTitle(a.GetIP())
				break
			case <-actionButton.ClickedCh:
				systray.Quit()
				break
			}
		}
	}()
}

func (a *app) GetIP() string {
	s, err := a.GetIpv4()
	if err != nil {
		return "0.0.0.0"
	}
	var output = fmt.Sprintf("%s:%v", s, a.Port)
	return output
}

// 复制`ipv4`到剪贴板
func (a *app) CopyIpv4() error {
	return clipboard.WriteAll(a.GetIP())
}

// 获取内网地址
func (a *app) GetIpv4() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// fmt.Println("address", address.String())
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				var ip = ipnet.IP.String()
				fmt.Println("内网ip: ", ip)
				return ip, nil
			}
		}
	}
	return "", errors.New("Can not find the client ip address!")
}

func (a *app) Ping() bool {
	return a.CtrServer.CheckServer()
}

func (a *app) Start() {
	a.CtrServer.RestartServer()
}

func (a *app) Stop() {
	a.CtrServer.StopServer()
}

func (a *app) Loop() {
	systray.Run(a.Init, func() {
		log.Println("已退出")
	})
}
