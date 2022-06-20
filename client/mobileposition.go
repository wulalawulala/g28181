package client

import (
	"g28181/device"

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

	return srv.SendMessageNoWithParas(keepaliveInfoStr)
}

func (srv *ServerOpt) MobileInfo(deviceID, longitude, latitude string) (sip.Response, error) {

	if deviceID == "" {
		deviceID = srv.ClientConfig.GB28181.GBID
	}

	mobilePositionInfoStr := device.MobilePositionInfoBuild(srv.GetSn(), deviceID, longitude, latitude)

	return srv.SendMessageNoWithParas(mobilePositionInfoStr)
}
