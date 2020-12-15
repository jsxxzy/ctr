package ctr

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/dustin/go-humanize"
	"github.com/jsxxzy/ctr/mag"
	"github.com/julienschmidt/httprouter"
	"github.com/kbinani/screenshot"
	"github.com/shirou/gopsutil/mem"
)

//go:generate go-bindata -o static.go -pkg ctr embed

// Server 服务
type Server struct {
	Port     int                // 端口
	router   *httprouter.Router // 路由
	httpWare *http.Server
}

// ClientInfo 客户端信息
type ClientInfo struct {
	OSname string                 // 名称
	Mem    *mem.VirtualMemoryStat // 内存使用
}

// callScreencast 截图
func callScreencast() ([]*image.RGBA, error) {
	var count = screenshot.NumActiveDisplays()
	var r []*image.RGBA
	for i := 0; i < count; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return nil, err
		}
		r = append(r, img)
		// fileName := fmt.Sprintf("%d_%dx%d.png", i, bounds.Dx(), bounds.Dy())
		// file, _ := os.Create(fileName)
		// defer file.Close()
		// png.Encode(file, img)
		// fmt.Printf("#%d : %v \"%s\"\n", i, bounds, fileName)
	}
	return r, nil
}

// GetHuman 获取格式化好的字符串
func (c *ClientInfo) GetHuman() string {
	var m = c.Mem
	var a, t = humanize.Bytes(m.Available), humanize.Bytes(m.Total)
	return fmt.Sprintf("%v/%v", a, t)
}

// New server
func New(port int) *Server {
	return &Server{
		Port:   port,
		router: httprouter.New(),
	}
}

// AddRouter 添加路由
func (s *Server) AddRouter(p, m string, handle httprouter.Handle) {
	s.router.Handle(m, p, handle)
}

// 将图片转为`base64`
//
// http://www.philo.top/2015/03/05/golangImageLibrary
//
// https://stackoverflow.com/a/42154713/10272586
func image2base64(img *image.RGBA) (r string) {
	emptyBuff := bytes.NewBuffer(nil)
	png.Encode(emptyBuff, img)
	r = base64.StdEncoding.EncodeToString(emptyBuff.Bytes())
	return
}

// Screencast 立刻截图
//
// 朋友们好啊, 刚才有个朋友给我发来几个截图, 我一看, 哦
// 原来是左佃, 有两个菜逼程序员, 一个写过一年, 一个写过两年
// 他们说, 额, 有一个说, 我在健身房练功颈椎写坏了, 陈老师你能不能教教我帮助治疗一下这个`bug`
// 我说可以, 我说你在健身房里`debug`没有, 他不服气, 我说小朋友, 你联调试试, 我一说他, 啪!!
// 就站起来了啊, 很快啊, 然后上来就是一个`docker`, 一个`alpine`镜像, 我全部防出去了啊, 防出去之后当然
// 就直接去世了, 然后运维就来了, 一拳就把我鼻子打骨折了
// 我说对不起, 对不起, 我不懂规矩, 我帮忙`debug`
// 两分钟以后, 就好了
// 但是他又说, 你这呲毛扎鬼的没用
// 诶, 我劝这两个老逼灯耗子尾汁, 好好反思
// 以后不要在找牌面了啊
// 谢谢朋友们!!!
func (s *Server) Screencast() ([]string, error) {
	l, e := callScreencast()
	if e != nil {
		return []string{}, e
	}
	var output []string
	for _, item := range l {
		output = append(output, image2base64(item))
	}
	return output, nil
}

// Screencasts 获取截图(base64s)
func (s *Server) Screencasts() []string {
	list, _ := s.Screencast()
	return list
}

// CInfo2 获取信息 => string(主机)
func (s *Server) CInfo2() string {
	var i = s.CInfo()
	return fmt.Sprintf("系统: %s\nCPU: %f\n内存: %s\n", i.OSname, i.Mem.UsedPercent, i.GetHuman())
}

// CInfo 获取信息(主机)
func (s *Server) CInfo() ClientInfo {
	dataInfo, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}
	cacheS, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return ClientInfo{
		OSname: cacheS,
		Mem:    dataInfo,
	}
}

// CShutdown 关机(主机)
func (s *Server) CShutdown() {
	mag.Shutdown()
}

// CReboot 重启(主机)
func (s *Server) CReboot() {
	mag.Reboot()
}

// CheckServer 判断服务是否在线
func (s *Server) CheckServer() bool {
	var p = s.Port
	var u = fmt.Sprintf("http://127.0.0.1:%v/ping", p)
	_, err := http.Get(u)
	return err == nil
}

// 生成监听地址
func (s *Server) getAddr() string {
	return fmt.Sprintf("0.0.0.0:%d", s.Port)
}

func (s *Server) initServer() {
	s.AddRouter("/ping", http.MethodGet, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Write([]byte("ok"))
	})
	s.AddRouter("/", http.MethodGet, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
				w.Write([]byte("未知错误"))
			}
		}()
		data, err := AssetString("embed/render.html")
		if err != nil {
			fmt.Print(err)
			panic("加载内嵌文件")
		}
		tpl := template.New("")
		_, err = tpl.Parse(data)
		if err != nil {
			fmt.Print(err)
			panic("格式化失败")
		}
		tpl.Execute(w, map[string]interface{}{
			"info":        s.CInfo2(),
			"screenshots": s.Screencasts(),
		})
	})
	s.AddRouter("/api/action", http.MethodPost, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		raw := buf.String()
		rawList := strings.Split(raw, "=")
		if len(rawList) <= 1 {
			w.Write([]byte("参数错误"))
			return
		}
		tt := rawList[1]
		T, err := strconv.Atoi(tt)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		switch T {
		case rebootAction: // 重启
			s.CReboot()
			break
		case shutdownAction: // 关机
			s.CShutdown()
			break
		}
		w.Write([]byte(tt))
	})
}

const (
	rebootAction = iota
	shutdownAction
)

var firstRunHTTP = true

// StartServer 启动服务
func (s *Server) StartServer() error {
	if firstRunHTTP {
		s.initServer()
		var p = s.getAddr()
		fmt.Printf("server run %s\n", p)
		s.httpWare = &http.Server{Addr: p, Handler: s.router}
		firstRunHTTP = false
	}
	err := s.httpWare.ListenAndServe()
	fmt.Println("err", err)
	return nil
}

// StopServer 暂停服务
func (s *Server) StopServer() bool {
	if s.httpWare == nil {
		return false
	}
	// return s.httpWare.Close() == nil
	return s.httpWare.Shutdown(context.Background()) == nil
}

// RestartServer 重启服务
//
// TODO
func (s *Server) RestartServer() {
	s.StopServer()
	s.StartServer()
}
