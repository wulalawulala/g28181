package main

import (
	"fmt"
	"g28181/client"
	"g28181/device"
	"time"

	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
)

func Read(g *device.GB28181Config) {

	g.ServerID = "44010200492000000001"
	g.Realm = "4401020049"
	// g.ServerIp = "192.168.1.164"
	g.ServerIp = "192.168.3.100"
	// g.ServerPort = 5060
	g.ServerPort = 5060
	g.UserName = "WVP_PWD"
	g.Password = "admin123"
	g.RegExpire = 3600
	g.KeepaliveInterval = 60
	g.MaxKeepaliveRetry = 5

	g.Transport = "UDP"
	// g.Transport = "TCP"
	g.LocalHost = "192.168.3.132"
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
func OnMessage(c *client.ClientConfigOption, req sip.Request, tx sip.ServerTransaction) {
	fmt.Println("GBID : ", c.GB28181.GBID)
	fmt.Printf("OnMessage : %s\n", req.String())

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
	go srv.Start()
	time.Sleep(time.Second)
	res, err := srv.Register()
	fmt.Println("Register err : ", err)
	if res != nil {
		fmt.Println("res : ", res.String())
	}
	select {}
}
