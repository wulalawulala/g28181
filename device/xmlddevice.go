package device

import (
	"encoding/xml"
	"strconv"
)

// MessageReceive 接收到的请求数据最外层，主要用来判断数据类型
type MessageReceive struct {
	CmdType string `xml:"CmdType"`
	SN      int    `xml:"SN"`
}

// Devices 摄像头信息
type Devices struct {
	DeviceID string `xml:"DeviceID" bson:"deviceid" json:"deviceid"` // DeviceID 设备编号
	// Name 设备名称
	Name         string `xml:"Name" bson:"name" json:"name"`
	Manufacturer string `xml:"Manufacturer" bson:"manufacturer" json:"manufacturer"`
	Model        string `xml:"Model" bson:"model" json:"model"`
	Owner        string `xml:"Owner" bson:"owner" json:"owner"`
	CivilCode    string `xml:"CivilCode" bson:"civilcode" json:"civilcode"`
	Block        string `xml:"Block" bson:"block" json:"block"`
	Address      string `xml:"Address" bson:"address" json:"address"`
	Parental     string `xml:"Parental" bson:"parental" json:"parental"`
	ParentID     string `xml:"ParentID" bson:"parentID" json:"parentID"`
	SafetyWay    string `xml:"SafetyWay" bson:"safetyway" json:"safetyway"`
	RegisterWay  string `xml:"RegisterWay" bson:"registerway" json:"registerway"`

	CertNum     string `xml:"CertNum" bson:"certnum" json:"certnum"`
	Certifiable string `xml:"Certifiable" bson:"certifiable" json:"certifiable"`
	ErrCode     string `xml:"ErrCode" bson:"errcode" json:"errcode"`
	EndTime     string `xml:"EndTime" bson:"endtime" json:"endtime"`
	IPAddress   string `xml:"IPAddress" bson:"ipaddress" json:"ipaddress"`
	Port        string `xml:"Port" bson:"port" json:"port"`

	Password string `xml:"Password" bson:"password" json:"password"`
	Status   string `xml:"Status" bson:"status" json:"status"` // Status 状态  on 在线

	Longitude string `xml:"Longitude" bson:"longitude" json:"longitude"`
	Latitude  string `xml:"Latitude" bson:"latitude" json:"latitude"`
}

// CatalogResponse 设备明细列表返回结构
type CatalogResponse struct {
	XMLName  xml.Name `xml:"Response"`
	CmdType  string   `xml:"CmdType"`
	SN       int      `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	SumNum   int      `xml:"SumNum"`

	Item []Devices `xml:"DeviceList>Item"`
}

// DeviceInfoResponse 主设备明细返回结构
type DeviceInfoResponse struct {
	XMLName      xml.Name `xml:"Response"`
	CmdType      string   `xml:"CmdType"`
	SN           int      `xml:"SN"`
	DeviceID     string   `xml:"DeviceID"`
	Result       string   `xml:"Result"`
	Manufacturer string   `xml:"Manufacturer"`
	Model        string   `xml:"Model"`
	Firmware     string   `xml:"Firmware"`
}

func (g *GB28181Config) DeviceInfoBuild(m MessageReceive, status string) string {
	var deviceInfoResponse DeviceInfoResponse
	deviceInfoResponse.CmdType = m.CmdType
	deviceInfoResponse.SN = m.SN
	deviceInfoResponse.DeviceID = g.GBID
	deviceInfoResponse.Result = status
	deviceInfoResponse.Manufacturer = g.Manufacturer
	deviceInfoResponse.Model = g.Model
	deviceInfoResponse.Firmware = g.Version
	bodyByte, err := XMLEncode(deviceInfoResponse)
	if err != nil {
		return ""
	}
	return string(bodyByte)
}

func (g *GB28181Config) CatalogBuild(m MessageReceive, Longitude, Latitude string) string {
	var catalogResponse CatalogResponse
	catalogResponse.CmdType = m.CmdType
	catalogResponse.SN = m.SN
	catalogResponse.DeviceID = g.GBID
	if g.Devices != nil {
		catalogResponse.SumNum = len(g.Devices)
		for i := 0; i < len(g.Devices); i++ {
			item := Devices{
				DeviceID:     g.Devices[i].DeviceID,
				Name:         g.Devices[i].Name,
				Manufacturer: g.Devices[i].Manufacturer,
				Model:        g.Devices[i].Model,
				Owner:        g.Devices[i].Owner,
				CivilCode:    g.Devices[i].CivilCode,
				Block:        "Block",
				Address:      g.Devices[i].Address,
				Parental:     g.Devices[i].Parental,
				ParentID:     g.Devices[i].ParentID,
				SafetyWay:    g.Devices[i].SafetyWay,
				RegisterWay:  g.Devices[i].RegisterWay,
				CertNum:      "CertNum",
				EndTime:      "2099-12-31T23:59:59",
				IPAddress:    g.LocalHost,
				Port:         strconv.Itoa(g.LocalSipPort),
				Status:       g.Devices[i].Status,
				Certifiable:  "0",
				ErrCode:      "400",
				Password:     "",
				Longitude:    Longitude,
				Latitude:     Latitude,
			}
			catalogResponse.Item = append(catalogResponse.Item, item)
		}
	}

	bodyByte, err := XMLEncode(catalogResponse)
	if err != nil {
		return ""
	}
	return string(bodyByte)
}
