package client

import "github.com/ghettovoice/gosip/sip"

func (srv *server) GetNewVia() sip.ViaHeader {
	port := sip.Port(srv.ClientConfig.GB28181.LocalSipPort)
	via := sip.ViaHeader{
		&sip.ViaHop{
			ProtocolName:    "SIP",
			ProtocolVersion: "2.0",
			Transport:       srv.ClientConfig.GB28181.Transport,
			Params:          sip.NewParams().Add("rport", sip.String{}).Add("branch", sip.String{Str: sip.GenerateBranch()}),
			Host:            srv.ClientConfig.GB28181.LocalHost,
			Port:            &port,
		},
	}
	return via
}

func (srv *server) GetSeq(method sip.RequestMethod) sip.CSeq {
	srv.ClientConfig.m.Lock()
	srv.ClientConfig.seqNo++
	cseq := sip.CSeq{
		SeqNo:      uint32(srv.ClientConfig.seqNo),
		MethodName: method,
	}
	srv.ClientConfig.m.Unlock()
	return cseq
}

func (srv *server) GetSn() int {
	srv.ClientConfig.msn.Lock()
	defer srv.ClientConfig.msn.Unlock()
	srv.ClientConfig.sn++
	sn := srv.ClientConfig.sn
	return sn
}