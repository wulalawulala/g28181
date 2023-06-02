package device

import (
	"encoding/xml"
	"time"
)

type MobilePositionReceive struct {
	CmdType string `xml:"CmdType"`
	SN      int    `xml:"SN"`

	DeviceID string `xml:"DeviceID"`
	Interval int    `xml:"Interval,omitempty"`
}

type MobilePositionInfo struct {
	XMLName xml.Name `xml:"Notify"`
	CmdType string   `xml:"CmdType"`
	SN      int      `xml:"SN"`

	DeviceID string `xml:"DeviceID"`
	// DeviceID string `xml:"TargetID"`

	Time      string `xml:"Time"`      //位置订阅-GPS时间
	Longitude string `xml:"Longitude"` //位置订阅-经度
	Latitude  string `xml:"Latitude"`  //位置订阅-维度
	Speed     string `xml:"Speed"`     //位置订阅-速度(km/h)(可选)
	Direction string `xml:"Direction"` //位置订阅-方向(取值为当前摄像头方向与正北方的顺时针夹角,取值范围0°~360°,单位:°)(可选)
	Altitude  string `xml:"Altitude"`  //位置订阅-海拔高度,单位:m(可选)
}

// KeepaliveInfo 设备明细列表返回结构
type KeepaliveInfo struct {
	XMLName  xml.Name `xml:"Notify"`
	CmdType  string   `xml:"CmdType"`
	SN       int      `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Status   string   `xml:"Status"`
}

func KeepaliveInfoBuild(sn int, deviceID, status string) string {
	var keepaliveInfo KeepaliveInfo
	keepaliveInfo.CmdType = "Keepalive"
	keepaliveInfo.SN = sn
	keepaliveInfo.DeviceID = deviceID
	keepaliveInfo.Status = status

	bodyByte, err := XMLEncode(keepaliveInfo)
	if err != nil {
		return ""
	}
	return string(bodyByte)
}

func MobilePositionInfoBuild(sn int, deviceID, longitude, latitude, speed string) string {
	var mobilePositionInfo MobilePositionInfo
	mobilePositionInfo.CmdType = "MobilePosition"
	mobilePositionInfo.SN = sn
	mobilePositionInfo.DeviceID = deviceID
	mobilePositionInfo.Time = time.Now().Format("2006-01-02T15:04:05")
	mobilePositionInfo.Longitude = longitude
	mobilePositionInfo.Latitude = latitude
	mobilePositionInfo.Speed = speed

	bodyByte, err := XMLEncode(mobilePositionInfo)
	if err != nil {
		return ""
	}
	return string(bodyByte)
}
