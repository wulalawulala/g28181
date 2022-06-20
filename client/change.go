package client

import (
	"strconv"
	"strings"
)

func (srv *ServerOpt) Start() error {

	laddr := "0.0.0.0:" + strconv.Itoa(int(srv.ClientConfig.GB28181.LocalSipPort))
	err := srv.Listen(strings.ToLower(srv.ClientConfig.GB28181.Transport), laddr)
	return err
}
