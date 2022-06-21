package client

import (
	"github.com/wulalawulala/g28181/device"

	"github.com/ghettovoice/gosip/sip"
)

func (srv *ServerOpt) Keepalive(deviceID, status string) (sip.Response, error) {

	if deviceID == "" {
		deviceID = srv.ClientConfig.GB28181.GBID
	}
	if status == "" {
		status = "OK"
	}

	keepaliveInfoStr := device.KeepaliveInfoBuild(srv.GetSn(), deviceID, status)

	res, err := srv.SendMessageNoWithParas(keepaliveInfoStr)

	if err == nil && res != nil {
		//设置nat的方法
		if via, ok := res.ViaHop(); ok {
			rport, ok1 := via.Params.Get("rport")
			received, ok2 := via.Params.Get("received")
			if ok1 && ok2 {
				if rport.String() != srv.ClientConfig.rport || received.String() != srv.ClientConfig.received {
					srv.ClientConfig.GetNewUaIpAddr(received.String(), rport.String())
				}
			}
		}
	}

	return res, err
}

func (srv *ServerOpt) MobileInfo(deviceID, longitude, latitude string) (sip.Response, error) {

	if deviceID == "" {
		deviceID = srv.ClientConfig.GB28181.GBID
	}

	mobilePositionInfoStr := device.MobilePositionInfoBuild(srv.GetSn(), deviceID, longitude, latitude)

	return srv.SendMessageNoWithParas(mobilePositionInfoStr)
}
