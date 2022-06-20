package main

import (
	"fmt"
	"g28181/client"
	"g28181/device"
	"g28181/sdp"
	"net/http"
	"time"

	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
)

func Read(g *device.GB28181Config) {

	g.ServerID = "44010200492000000001"
	g.Realm = "4401020049"
	g.ServerIp = "192.168.1.110"
	// g.ServerPort = 5060
	g.ServerPort = 5060
	g.UserName = "WVP_PWD"
	g.Password = "admin123"
	g.RegExpire = 3600
	g.KeepaliveInterval = 60
	g.MaxKeepaliveRetry = 5

	g.Transport = "UDP"
	// g.Transport = "TCP"
	g.LocalHost = "192.168.1.164"
	// g.LocalHost = "" //本地ip
	g.LocalSipPort = 5060

	g.GBID = "11223344556677"
	g.Version = "00.00.01"
	g.Devices = make([]device.DeviceInfo, 1)

	g.Devices[0].DeviceID = g.GBID
	g.Devices[0].Name = "yl_ipc"
	g.Devices[0].Manufacturer = "xmrbi"
	g.Devices[0].Model = "xmrbi"
	g.Devices[0].Owner = "Owner"
	g.Devices[0].CivilCode = "CivilCode"
	g.Devices[0].Address = g.LocalHost
	g.Devices[0].Parental = "0"
	g.Devices[0].ParentID = g.GBID
	g.Devices[0].SafetyWay = "0"
	g.Devices[0].RegisterWay = "1"
	g.Devices[0].Secrecy = "1"
	g.Devices[0].Status = "ON"

}

func OnMessage(s *client.ServerOpt, req sip.Request, tx sip.ServerTransaction) {
	// fmt.Println("GBID : ", s)
	fmt.Printf("OnMessage : %s\n", req.String())
	//应答ok必不可少
	tx.Respond(sip.NewResponseFromRequest(req.MessageID(), req, http.StatusOK, http.StatusText(http.StatusOK), ""))

	from, ok := req.From()
	if !ok {
		return
	}
	serverID := from.Address.User().String()
	if serverID != s.ClientConfig.GB28181.ServerID {
		return
	}
	body := req.Body()
	message := &device.MessageReceive{}
	if err := device.XMLDecode([]byte(body), message); err != nil {
		return
	}
	//应答
	switch message.CmdType {
	case "Catalog":
		var source, destination, transport string
		source = req.Source()
		destination = req.Destination()
		transport = req.Transport()
		r_body := s.ClientConfig.GB28181.CatalogBuild(*message, "", "")
		go func() {
			s.SendMessage(source, destination, transport, r_body)
		}()
	case "DeviceInfo":
		var source, destination, transport string
		source = req.Source()
		destination = req.Destination()
		transport = req.Transport()
		r_body := s.ClientConfig.GB28181.DeviceInfoBuild(*message, "OK")
		go func() {
			s.SendMessage(source, destination, transport, r_body)
		}()
	}

}

func OnInvite(s *client.ServerOpt, req sip.Request, tx sip.ServerTransaction) {
	fmt.Printf("OnInvite : %s\n", req.String())
	tx.Respond(sip.NewResponseFromRequest(req.MessageID(), req, http.StatusContinue, "Trying", ""))

	from, ok := req.From()
	if !ok {
		return
	}
	serverID := from.Address.User().String()
	if serverID != s.ClientConfig.GB28181.ServerID {
		return
	}

	var ssdp string
	//解析sdp
	body := req.Body()
	session, err := sdp.ParseString(body)
	if err == nil {
		if len(session.Media) > 0 {
			ssdp = device.BuildLocalSdp(s.ClientConfig.GB28181.GBID, s.ClientConfig.GB28181.LocalHost, 0, session.Media[0].SSRC)
			// 这里获取到了session.Media[0].SSRC 发送rtp的ssrc标识符
			// session.Connection.Address 发送rtp的ip地址
			// session.Media[0].Port 发送rtp的端口
			// 开始流传输,自定义操作
		}
	}
	s.RespondSdp(req, tx, sdp.ContentType, ssdp)
}
func OnBye(s *client.ServerOpt, req sip.Request, tx sip.ServerTransaction) {
	fmt.Printf("OnBye : %s\n", req.String())
	tx.Respond(sip.NewResponseFromRequest(req.MessageID(), req, http.StatusOK, http.StatusText(http.StatusOK), ""))

	from, ok := req.From()
	if !ok {
		return
	}
	serverID := from.Address.User().String()
	if serverID != s.ClientConfig.GB28181.ServerID {
		return
	}
	//关闭流传输,自定义操作

}

func main() {
	logger := log.NewDefaultLogrusLogger()
	var gb28181Config device.GB28181Config
	Read(&gb28181Config)
	srvconf := client.ServerConfig{
		ClientConfig: client.ClientConfigOption{
			GB28181: gb28181Config,
		},
	}

	srv := client.NewServer(srvconf, nil, nil, logger)
	if srv == nil {
		return
	}
	srv.OnRequest(sip.MESSAGE, OnMessage)
	srv.OnRequest(sip.INVITE, OnInvite)
	srv.OnRequest(sip.BYE, OnBye)
	go srv.Start()
	time.Sleep(time.Second)
	res, err := srv.Register()
	fmt.Println("Register err : ", err)
	if res != nil {
		fmt.Println("res : ", res.String())
	}
	select {}
}
