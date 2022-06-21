package client

import (
	"context"
	"errors"
	"time"

	"github.com/wulalawulala/g28181/device"

	"github.com/ghettovoice/gosip/sip"
)

func (srv *ServerOpt) Register() (sip.Response, error) {
	if !srv.running.IsSet() {
		return nil, errors.New("please Start srv")
	}
	callId := sip.CallID(device.RandNumString(10))
	cseq := srv.GetSeq(sip.REGISTER)
	expires := sip.Expires(srv.ClientConfig.GB28181.RegExpire)
	contentlength := sip.ContentLength(0)
	maxForwards := sip.MaxForwards(70)
	req := sip.NewRequest(
		"",
		sip.REGISTER,
		srv.ClientConfig.ServerIp.Uri,
		DefaultsipVersion,
		[]sip.Header{
			srv.GetNewVia(),
			srv.ClientConfig.UaRealmAddr.AsFromHeader(),
			srv.ClientConfig.UaRealmAddr.AsToHeader(),
			&callId,
			&cseq,
			srv.ClientConfig.UaIpAddr.AsContactHeader(),
			&expires,
			&maxForwards,
			&contentlength,
		},
		"",
		nil,
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var authorizer sip.Authorizer = &sip.DefaultAuthorizer{
		User: sip.String{
			Str: srv.ClientConfig.GB28181.UserName,
		},
		Password: sip.String{
			Str: srv.ClientConfig.GB28181.Password,
		},
	}
	rsp, err := srv.RequestWithContext(ctx, req, WithAuthorizer(authorizer))

	if rsp != nil {
		//设置nat的方法
		if via, ok := rsp.ViaHop(); ok {
			rport, ok1 := via.Params.Get("rport")
			received, ok2 := via.Params.Get("received")
			if ok1 && ok2 {
				srv.ClientConfig.GetNewUaIpAddr(received.String(), rport.String())
			}
		}
	}
	return rsp, err
}
