package client

import (
	"net/http"

	"github.com/ghettovoice/gosip/sip"
)

func (srv *ServerOpt) RespondSdp(req sip.Request, tx sip.ServerTransaction, contentType, ssdp string) error {
	res := sip.NewResponseFromRequest("", req, http.StatusOK, http.StatusText(http.StatusOK), ssdp)

	to := srv.ClientConfig.UaRealmAddr.AsToHeader()
	res.ReplaceHeaders(to.Name(), []sip.Header{to})
	cseq := srv.GetSeq(sip.INVITE)
	res.ReplaceHeaders(cseq.Name(), []sip.Header{&cseq})
	res.AppendHeader(srv.ClientConfig.NewUaIpAddr.AsContactHeader())
	res.AppendHeader(&sip.GenericHeader{HeaderName: "Content-Type", Contents: contentType})

	err := tx.Respond(res)
	return err
}
