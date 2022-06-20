package client

import (
	"context"
	"g28181/device"
	"time"

	"github.com/ghettovoice/gosip/sip"
)

func (srv *ServerOpt) SendMessage(source, destination, transport string, body string) (sip.Response, error) {
	callId := sip.CallID(device.RandNumString(10))
	cseq := srv.GetSeq(sip.MESSAGE)
	maxForwards := sip.MaxForwards(70)

	req := sip.NewRequest(
		"",
		sip.MESSAGE,
		srv.ClientConfig.ServerIp.Uri,
		DefaultsipVersion,
		[]sip.Header{
			srv.GetNewVia(),
			srv.ClientConfig.UaIpAddr.AsFromHeader(),
			srv.ClientConfig.ServeIpAddr.AsToHeader(),
			srv.ClientConfig.NewUaIpAddr.AsContactHeader(),
			&callId,
			&cseq,
			&maxForwards,
		},
		"",
		nil,
	)
	req.SetDestination(source)
	req.SetSource(destination)
	req.SetTransport(transport)
	if body != "" {
		contentType := sip.ContentType("Application/MANSCDP+xml")
		req.AppendHeader(&contentType)
		req.SetBody(body, true)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	rsp, err := srv.RequestWithContext(ctx, req)
	return rsp, err
}
