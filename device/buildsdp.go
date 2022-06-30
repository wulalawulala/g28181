package device

import (
	"time"

	"github.com/wulalawulala/g28181/sdp"
)

//port填0,填其他数值发送不到服务器(目前无法解决)
func BuildLocalSdp(userName string, host string, port int, ssrc string) string {
	sdpOpt := &sdp.Session{
		Origin: &sdp.Origin{
			Username: userName,
			Address:  host,
		},
		Timing: &sdp.Timing{
			Start: time.Time{},
			Stop:  time.Time{},
		},
		Name: "Playback",
		Connection: &sdp.Connection{
			Type:    "ip4",
			Address: host,
		},
		//Bandwidth: []*sdp.Bandwidth{{Type: "AS", Value: 117}},
		Media: []*sdp.Media{
			{
				//Bandwidth: []*sdp.Bandwidth{{Type: "TIAS", Value: 96000}},
				// Connection: []*sdp.Connection{{Address: host}},
				Mode:  sdp.SendOnly,
				Type:  "video",
				Port:  port,
				Proto: "RTP/AVP",
				Format: []*sdp.Format{
					// {Payload: 98, Name: "H264", ClockRate: 90000},
					// {Payload: 97, Name: "MPEG4", ClockRate: 90000},
					{Payload: 96, Name: "PS", ClockRate: 90000},
				},
				SSRC: ssrc,
			},
		},
	}

	return sdpOpt.String()
}

func BuildLocalSdpWithTime(userName string, host string, port int, ssrc string, s, e time.Time) string {
	sdpOpt := &sdp.Session{
		Origin: &sdp.Origin{
			Username: userName,
			Address:  host,
		},
		Timing: &sdp.Timing{
			Start: s,
			Stop:  e,
		},
		Name: "Playback",
		Connection: &sdp.Connection{
			Type:    "ip4",
			Address: host,
		},
		//Bandwidth: []*sdp.Bandwidth{{Type: "AS", Value: 117}},
		Media: []*sdp.Media{
			{
				//Bandwidth: []*sdp.Bandwidth{{Type: "TIAS", Value: 96000}},
				// Connection: []*sdp.Connection{{Address: host}},
				Mode:  sdp.SendOnly,
				Type:  "video",
				Port:  port,
				Proto: "RTP/AVP",
				Format: []*sdp.Format{
					// {Payload: 98, Name: "H264", ClockRate: 90000},
					// {Payload: 97, Name: "MPEG4", ClockRate: 90000},
					{Payload: 96, Name: "PS", ClockRate: 90000},
				},
				SSRC: ssrc,
			},
		},
	}

	return sdpOpt.String()
}

//port填0,填其他数值发送不到服务器(目前无法解决)
func BuildLocalSdpTSWithTime(userName string, host string, port int, ssrc string, s, e time.Time) string {
	sdpOpt := &sdp.Session{
		Origin: &sdp.Origin{
			Username: userName,
			Address:  host,
		},
		Timing: &sdp.Timing{
			Start: s,
			Stop:  e,
		},
		Name: "Playback",
		Connection: &sdp.Connection{
			Type:    "ip4",
			Address: host,
		},
		//Bandwidth: []*sdp.Bandwidth{{Type: "AS", Value: 117}},
		Media: []*sdp.Media{
			{
				//Bandwidth: []*sdp.Bandwidth{{Type: "TIAS", Value: 96000}},
				// Connection: []*sdp.Connection{{Address: host}},
				Mode:  sdp.SendOnly,
				Type:  "video",
				Port:  port,
				Proto: "RTP/AVP",
				Format: []*sdp.Format{
					// {Payload: 98, Name: "H264", ClockRate: 90000},
					{Payload: 97, Name: "MPEG4", ClockRate: 90000},
					// {Payload: 96, Name: "PS", ClockRate: 90000},
				},
				SSRC: ssrc,
			},
		},
	}

	return sdpOpt.String()
}
