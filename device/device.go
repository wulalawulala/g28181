package device

type DeviceInfo struct {
	Text         string `xml:",chardata"`                        //
	DeviceID     string `xml:"DeviceID" json:"deviceID"`         //子设备国标ID
	Name         string `xml:"Name" json:"name""`                //子设备名称
	Manufacturer string `xml:"Manufacturer" json:"manufacturer"` //子设备厂商
	Model        string `xml:"Model" json:"model"`               //子设备model
	Owner        string `xml:"Owner" json:"owner"`               //
	CivilCode    string `xml:"CivilCode" json:"civilCode"`       //
	Address      string `xml:"Address" json:"address"`           //子设备ip地址
	Parental     string `xml:"Parental" json:"parental"`         //
	ParentID     string `xml:"ParentID" json:"ParentID"`         //
	SafetyWay    string `xml:"SafetyWay" json:"safeWay"`         //
	RegisterWay  string `xml:"RegisterWay" json:"registerWay"`   //
	Secrecy      string `xml:"Secrecy" json:"secrecy"`           //
	Status       string `xml:"Status" json:"status"`             //子设备状态
}

type GB28181Config struct {
	ServerID          string       `json:"serverID"`          //服务器 id, 默认 34020000002000000001
	Realm             string       `json:"realm"`             //服务器域, 默认 3402000000
	ServerIp          string       `json:"serverIp"`          //服务器公网IP
	ServerPort        uint16       `json:"serverPort"`        //服务器公网端口
	UserName          string       `json:"userName"`          //服务器账号
	Password          string       `json:"password"`          //服务器密码
	RegExpire         int          `json:"regExpire"`         //注册有效期，单位秒，默认 3600
	KeepaliveInterval int          `json:"keepaliveInterval"` //keepalive 心跳时间
	MaxKeepaliveRetry int          `json:"maxKeepaliveRetry"` //keeplive超时次数(超时之后发送重新发送reg)
	Transport         string       `json:"transport"`         //传输层协议(目前只支持udp,tcp)
	LocalHost         string       `json:"localHost"`         //本地的ip地址,如果是空,则自动获取
	LocalSipPort      int          `json:"localSipPort"`      //本地的端口
	GBID              string       `json:"gbID"`              //设备国标ID
	Version           string       `json:"Version"`
	Manufacturer      string       `json:"Manufacturer"` //设备厂商
	Model             string       `json:"Model"`        //设备model
	Devices           []DeviceInfo `json:"devices"`      //从设备地址
}
