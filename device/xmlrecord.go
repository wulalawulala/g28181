package device

type MessageRecordInfoReceive struct {
	CmdType string `xml:"CmdType"`
	SN      int    `xml:"SN"`

	DeviceID  string `xml:"DeviceID"`
	StartTime string `xml:"StartTime"`
	EndTime   string `xml:"EndTime"`
	Secrecy   int    `xml:"Secrecy"`
	Type      string `xml:"Type"`
}

type RecordItem struct {
	// DeviceID 设备编号
	DeviceID string `xml:"DeviceID" bson:"DeviceID" json:"DeviceID"`
	// Name 设备名称
	Name      string `xml:"Name" bson:"Name" json:"Name"`
	FilePath  string `xml:"FilePath" bson:"FilePath" json:"FilePath"`
	Address   string `xml:"Address" bson:"Address" json:"Address"`
	StartTime string `xml:"StartTime" bson:"StartTime" json:"StartTime"`
	EndTime   string `xml:"EndTime" bson:"EndTime" json:"EndTime"`
	Secrecy   int    `xml:"Secrecy" bson:"Secrecy" json:"Secrecy"`
	Type      string `xml:"Type" bson:"Type" json:"Type"`
}

type RecordLists struct {
	Num  int          `xml:"Num,attr"`
	Item []RecordItem `xml:"Item"`
}

type MessageRecordInfoRsp struct {
	CmdType  string `xml:"CmdType"`
	SN       int    `xml:"SN"`
	DeviceID string `xml:"DeviceID"`
	SumNum   int    `xml:"SumNum"`
	// Item     []RecordItem `xml:"RecordList>Item"`
	Records RecordLists `xml:"RecordList"`
}

//2006-01-02T15:04:05

func (m *MessageRecordInfoRsp) BuildString() string {

	bodyByte, err := XMLEncode(m)
	if err != nil {
		return ""
	}
	return string(bodyByte)
}

type MessageRecordInfoRspV2 struct {
	CmdType  string       `xml:"CmdType"`
	SN       int          `xml:"SN"`
	DeviceID string       `xml:"DeviceID"`
	SumNum   int          `xml:"SumNum"`
	Item     []RecordItem `xml:"RecordList>Item"`
	// Records RecordLists `xml:"RecordList"`
}

func (m *MessageRecordInfoRspV2) BuildString() string {

	bodyByte, err := XMLEncode(m)
	if err != nil {
		return ""
	}
	return string(bodyByte)
}
