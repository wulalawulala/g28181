package client

import (
	"fmt"
	"g28181/device"
	"sync"

	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/sip/parser"
)

type ClientConfigOption struct {
	GB28181 device.GB28181Config

	UaRealmAddr *sip.Address //sip:%s@%s sip:31011500991320000343@4401020049 携带tag Params: sip.NewParams().Add("tag", sip.String{Str: utils.RandNumString(9)})
	UaIpAddr    *sip.Address //sip:%s@%s:%d sip:31011500991320000343@192.168.3.105:5060
	NewUaIpAddr *sip.Address //sip:%s@%s:%d sip:31011500991320000343@received:rport //返回的via

	ServerIp        *sip.Address //dst ip port   FHost: "192.168.3.100", FPort: 5060
	ServerRealmAddr *sip.Address //sip:%s@%s sip:44010200492000000001@4401020049
	ServeIpAddr     *sip.Address //sip:%s@%s sip:44010200492000000001@192.168.3.100:5060

	m     *sync.Mutex //统计计数的锁
	seqNo uint32      //累计计数

	msn *sync.Mutex //sn计数的锁
	sn  int         //sn计数
}

func (c *ClientConfigOption) GetUaOption() error {
	uaRealmUrl, err := parser.ParseUri(fmt.Sprintf("sip:%s@%s", c.GB28181.GBID, c.GB28181.Realm))
	if err != nil {
		return err
	}
	c.UaRealmAddr = &sip.Address{
		Uri:    uaRealmUrl,
		Params: sip.NewParams().Add("tag", sip.String{Str: device.RandNumString(9)}),
	}

	uaIpUrl, err := parser.ParseUri(fmt.Sprintf("sip:%s@%s:%d", c.GB28181.GBID, c.GB28181.LocalHost, c.GB28181.LocalSipPort))
	if err != nil {
		return err
	}
	c.UaIpAddr = &sip.Address{
		Uri: uaIpUrl,
	}

	return nil
}

func (c *ClientConfigOption) GetNewUaIpAddr(received, rport string) error {
	c.NewUaIpAddr = c.UaIpAddr
	newUaIpAddr, err := parser.ParseUri(fmt.Sprintf("sip:%s@%s:%s", c.GB28181.GBID, received, rport))
	if err != nil {
		return err
	}
	c.NewUaIpAddr = &sip.Address{
		Uri: newUaIpAddr,
	}
	return nil
}

func (c *ClientConfigOption) GetServerOption() error {
	port := sip.Port(c.GB28181.ServerPort)
	c.ServerIp = &sip.Address{
		Uri: &sip.SipUri{
			FHost: c.GB28181.ServerIp,
			FPort: &port,
		},
	}
	serverRealmUrl, err := parser.ParseUri(fmt.Sprintf("sip:%s@%s", c.GB28181.ServerID, c.GB28181.Realm))
	if err != nil {
		return err
	}
	c.ServerRealmAddr = &sip.Address{
		Uri: serverRealmUrl,
	}
	serveIpUrl, err := parser.ParseUri(fmt.Sprintf("sip:%s@%s:%d", c.GB28181.ServerID, c.GB28181.ServerIp, c.GB28181.ServerPort))
	if err != nil {
		return err
	}
	c.ServeIpAddr = &sip.Address{
		Uri: serveIpUrl,
	}

	return nil
}
