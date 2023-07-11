package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/transaction"
	"github.com/ghettovoice/gosip/transport"
	"github.com/ghettovoice/gosip/util"
	"github.com/wulalawulala/g28181/device"

	"github.com/tevino/abool"
)

// RequestHandler is a callback that will be called on the incoming request
// of the certain method
// tx argument can be nil for 2xx ACK request
type RequestHandler func(srv *ServerOpt, req sip.Request, tx sip.ServerTransaction)

type Server interface {
	Start() error
	Shutdown()

	Listen(network, addr string, options ...transport.ListenOption) error
	Send(msg sip.Message) error

	Request(req sip.Request) (sip.ClientTransaction, error)
	RequestWithContext(
		ctx context.Context,
		request sip.Request,
		options ...RequestWithContextOption,
	) (sip.Response, error)
	OnRequest(method sip.RequestMethod, handler RequestHandler) error

	Respond(res sip.Response) (sip.ServerTransaction, error)
	RespondOnRequest(
		request sip.Request,
		status sip.StatusCode,
		reason, body string,
		headers []sip.Header,
	) (sip.ServerTransaction, error)

	Register() (sip.Response, error)
	SendMessageNoWithParas(body string) (sip.Response, error)

	Keepalive(deviceID, status string) (sip.Response, error)
	MobileInfo(deviceID, longitude, latitude, speed string) (sip.Response, error)
	GetConfig() *ClientConfigOption
	GetServerOpt() *ServerOpt
}

type TransportLayerFactory func(
	ip net.IP,
	dnsResolver *net.Resolver,
	msgMapper sip.MessageMapper,
	logger log.Logger,
) transport.Layer

type TransactionLayerFactory func(tpl sip.Transport, logger log.Logger) transaction.Layer

// ServerConfig describes available options
type ServerConfig struct {
	// Public IP address or domain name, if empty auto resolved IP will be used.
	Host string
	// Dns is an address of the public DNS ServerOpt to use in SRV lookup.
	Dns        string
	Extensions []string
	MsgMapper  sip.MessageMapper
	UserAgent  string

	ClientConfig ClientConfigOption
}

// Server is a SIP ServerOpt
type ServerOpt struct {
	running         abool.AtomicBool
	tp              transport.Layer
	tx              transaction.Layer
	host            string
	ip              net.IP
	hwg             *sync.WaitGroup
	hmu             *sync.RWMutex
	requestHandlers map[sip.RequestMethod]RequestHandler
	extensions      []string
	userAgent       string

	ClientConfig ClientConfigOption //一些配置参数定义

	log log.Logger
}

// NewServer creates new instance of SIP ServerOpt.
func NewServer(
	config ServerConfig,
	tpFactory TransportLayerFactory,
	txFactory TransactionLayerFactory,
) Server {
	if tpFactory == nil {
		tpFactory = transport.NewLayer
	}
	if txFactory == nil {
		txFactory = transaction.NewLayer
	}
	// if config.ClientConfig.L == nil {
	config.ClientConfig.L = log.NewDefaultLogrusLogger().WithPrefix("gosip.Server")
	// }

	if config.Host == "" {
		config.Host = config.ClientConfig.GB28181.LocalHost
	}
	if config.ClientConfig.crwm == nil {
		config.ClientConfig.crwm = new(sync.RWMutex)
	}
	if config.ClientConfig.m == nil {
		config.ClientConfig.m = new(sync.Mutex)
	}
	if config.ClientConfig.msn == nil {
		config.ClientConfig.msn = new(sync.Mutex)
	}
	//初始配置初始化
	config.ClientConfig.params =
		sip.NewParams().Add("tag", sip.String{Str: device.RandNumString(9)})
	config.ClientConfig.GetUaOption()
	config.ClientConfig.GetServerOption()

	var host string
	var ip net.IP
	if config.Host != "" {
		host = config.Host
		if addr, err := net.ResolveIPAddr("ip", host); err == nil {
			ip = addr.IP
		} else {
			config.ClientConfig.L.Panicf("resolve host IP failed: %s", err)
		}
	} else {
		if v, err := util.ResolveSelfIP(); err == nil {
			ip = v
			host = v.String()
		} else {
			config.ClientConfig.L.Panicf("resolve host IP failed: %s", err)
		}
	}

	var dnsResolver *net.Resolver
	if config.Dns != "" {
		dnsResolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "udp", config.Dns)
			},
		}
	} else {
		dnsResolver = net.DefaultResolver
	}

	var extensions []string
	if config.Extensions != nil {
		extensions = config.Extensions
	}

	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = "GoSIP"
	}

	srv := &ServerOpt{
		host:            host,
		ip:              ip,
		hwg:             new(sync.WaitGroup),
		hmu:             new(sync.RWMutex),
		requestHandlers: make(map[sip.RequestMethod]RequestHandler),
		extensions:      extensions,
		userAgent:       userAgent,
		ClientConfig:    config.ClientConfig,
	}
	srv.log = config.ClientConfig.L.WithFields(log.Fields{
		"sip_ServerOpt_ptr": fmt.Sprintf("%p", srv),
	})
	srv.tp = tpFactory(ip, dnsResolver, config.MsgMapper, srv.Log())
	sipTp := &sipTransport{
		tpl: srv.tp,
		srv: srv,
	}
	srv.tx = txFactory(sipTp, log.AddFieldsFrom(srv.Log(), srv.tp))

	srv.running.Set()
	go srv.serve()

	return srv
}

func (srv *ServerOpt) Log() log.Logger {
	return srv.log
}

// ListenAndServe starts serving listeners on the provided address
func (srv *ServerOpt) Listen(network string, listenAddr string, options ...transport.ListenOption) error {
	return srv.tp.Listen(network, listenAddr, options...)
}

func (srv *ServerOpt) serve() {
	defer srv.Shutdown()

	for {
		select {
		case tx, ok := <-srv.tx.Requests():
			if !ok {
				return
			}
			srv.hwg.Add(1)
			go srv.handleRequest(tx.Origin(), tx)
		case ack, ok := <-srv.tx.Acks():
			if !ok {
				return
			}
			srv.hwg.Add(1)
			go srv.handleRequest(ack, nil)
		case response, ok := <-srv.tx.Responses():
			if !ok {
				return
			}

			logger := srv.Log().WithFields(response.Fields())
			logger.Warn("received not matched response")

			// FIXME do something with this?
		case err, ok := <-srv.tx.Errors():
			if !ok {
				return
			}

			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				srv.Log().Debugf("received SIP transaction error: %s", err)
			} else {
				srv.Log().Errorf("received SIP transaction error: %s", err)
			}
		case err, ok := <-srv.tp.Errors():
			if !ok {
				return
			}

			var ferr *sip.MalformedMessageError
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				srv.Log().Debugf("received SIP transport error: %s", err)
			} else if errors.As(err, &ferr) {
				srv.Log().Warnf("received SIP transport error: %s", err)
			} else {
				srv.Log().Errorf("received SIP transport error: %s", err)
			}
		}
	}
}

func (srv *ServerOpt) handleRequest(req sip.Request, tx sip.ServerTransaction) {
	defer srv.hwg.Done()

	logger := srv.Log().WithFields(req.Fields())
	logger.Debug("routing incoming SIP request...")

	srv.hmu.RLock()
	handler, ok := srv.requestHandlers[req.Method()]
	srv.hmu.RUnlock()

	if !ok {
		logger.Warn("SIP request handler not found")

		// ACK request doesn't have any transaction, so just skip this step
		if tx != nil {
			go func(tx sip.ServerTransaction, logger log.Logger) {
				for {
					select {
					case <-srv.tx.Done():
						return
					case err, ok := <-tx.Errors():
						if !ok {
							return
						}

						logger.Warnf("error from SIP ServerOpt transaction %s: %s", tx, err)
					}
				}
			}(tx, logger)
		}

		// ACK request doesn't require any response, so just skip this step
		if !req.IsAck() {
			res := sip.NewResponseFromRequest("", req, 405, "Method Not Allowed", "")
			if _, err := srv.Respond(res); err != nil {
				logger.Errorf("respond '405 Method Not Allowed' failed: %s", err)
			}
		}

		return
	}

	go handler(srv, req, tx)
}

// Send SIP message
func (srv *ServerOpt) Request(req sip.Request) (sip.ClientTransaction, error) {
	if !srv.running.IsSet() {
		return nil, fmt.Errorf("can not send through stopped ServerOpt")
	}

	return srv.tx.Request(srv.prepareRequest(req))
}

func (srv *ServerOpt) RequestWithContext(
	ctx context.Context,
	request sip.Request,
	options ...RequestWithContextOption,
) (sip.Response, error) {
	return srv.requestWithContext(ctx, request, 1, options...)
}

func (srv *ServerOpt) requestWithContext(
	ctx context.Context,
	request sip.Request,
	attempt int,
	options ...RequestWithContextOption,
) (sip.Response, error) {
	tx, err := srv.Request(request)
	if err != nil {
		return nil, err
	}

	optionsHash := &RequestWithContextOptions{}
	for _, opt := range options {
		opt.ApplyRequestWithContext(optionsHash)
	}

	txResponses := tx.Responses()
	txErrs := tx.Errors()
	responses := make(chan sip.Response, 1)
	errs := make(chan error, 1)
	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()

		select {
		case <-done:
		case <-ctx.Done():
			if err := tx.Cancel(); err != nil {
				srv.Log().Error("cancel transaction failed", log.Fields{
					"transaction_key": tx.Key(),
				})
			}
		}
	}()
	go func() {
		defer func() {
			close(done)
			wg.Done()
		}()

		var lastResponse sip.Response

		previousMessages := make([]sip.Response, 0)
		previousResponsesStatuses := make(map[string]bool)
		getKey := func(res sip.Response) string {
			return fmt.Sprintf("%d %s", res.StatusCode(), res.Reason())
		}

		for {
			select {
			case err, ok := <-txErrs:
				if !ok {
					txErrs = nil
					// errors chan was closed
					// we continue to pull responses until close
					continue
				}
				errs <- err
				return
			case response, ok := <-txResponses:
				if !ok {
					if lastResponse != nil {
						lastResponse.SetPrevious(previousMessages)
					}
					errs <- sip.NewRequestError(487, "Request Terminated", request, lastResponse)
					return
				}

				response = sip.CopyResponse(response)
				lastResponse = response

				if optionsHash.ResponseHandler != nil {
					optionsHash.ResponseHandler(response, request)
				}

				if response.IsProvisional() {
					if _, ok := previousResponsesStatuses[getKey(response)]; !ok {
						previousMessages = append(previousMessages, response)
						previousResponsesStatuses[getKey(response)] = true
					}

					continue
				}

				// success
				if response.IsSuccess() {
					response.SetPrevious(previousMessages)
					responses <- response

					go func() {
						for response := range tx.Responses() {
							if optionsHash.ResponseHandler != nil {
								optionsHash.ResponseHandler(response, request)
							}
						}
					}()

					return
				}

				// unauth request
				needAuth := (response.StatusCode() == 401 || response.StatusCode() == 407) && attempt < 2
				if needAuth && optionsHash.Authorizer != nil {
					if err := optionsHash.Authorizer.AuthorizeRequest(request, response); err != nil {
						errs <- err

						return
					}

					if response, err := srv.requestWithContext(ctx, request, attempt+1, options...); err == nil {
						responses <- response
					} else {
						errs <- err
					}

					return
				}

				// failed request
				response.SetPrevious(previousMessages)
				errs <- sip.NewRequestError(uint(response.StatusCode()), response.Reason(), request, response)

				return
			}
		}
	}()

	var res sip.Response
	select {
	case err = <-errs:
	case res = <-responses:
	}

	wg.Wait()

	return res, err
}

func (srv *ServerOpt) prepareRequest(req sip.Request) sip.Request {
	srv.appendAutoHeaders(req)

	return req
}

func (srv *ServerOpt) Respond(res sip.Response) (sip.ServerTransaction, error) {
	if !srv.running.IsSet() {
		return nil, fmt.Errorf("can not send through stopped ServerOpt")
	}

	return srv.tx.Respond(srv.prepareResponse(res))
}

func (srv *ServerOpt) RespondOnRequest(
	request sip.Request,
	status sip.StatusCode,
	reason, body string,
	headers []sip.Header,
) (sip.ServerTransaction, error) {
	response := sip.NewResponseFromRequest("", request, status, reason, body)
	for _, header := range headers {
		response.AppendHeader(header)
	}

	tx, err := srv.Respond(response)
	if err != nil {
		return nil, fmt.Errorf("respond '%d %s' failed: %w", response.StatusCode(), response.Reason(), err)
	}

	return tx, nil
}

func (srv *ServerOpt) Send(msg sip.Message) error {
	if !srv.running.IsSet() {
		return fmt.Errorf("can not send through stopped ServerOpt")
	}

	switch m := msg.(type) {
	case sip.Request:
		msg = srv.prepareRequest(m)
	case sip.Response:
		msg = srv.prepareResponse(m)
	}

	return srv.tp.Send(msg)
}

func (srv *ServerOpt) prepareResponse(res sip.Response) sip.Response {
	srv.appendAutoHeaders(res)

	return res
}

// Shutdown gracefully shutdowns SIP ServerOpt
func (srv *ServerOpt) Shutdown() {
	if !srv.running.IsSet() {
		return
	}
	srv.running.UnSet()
	// stop transaction layer
	srv.tx.Cancel()
	<-srv.tx.Done()
	// stop transport layer
	srv.tp.Cancel()
	<-srv.tp.Done()
	// wait for handlers
	srv.hwg.Wait()
}

// OnRequest registers new request callback
func (srv *ServerOpt) OnRequest(method sip.RequestMethod, handler RequestHandler) error {
	srv.hmu.Lock()
	srv.requestHandlers[method] = handler
	srv.hmu.Unlock()

	return nil
}

func (srv *ServerOpt) appendAutoHeaders(msg sip.Message) {
	autoAppendMethods := map[sip.RequestMethod]bool{
		sip.INVITE:   true,
		sip.REGISTER: true,
		sip.OPTIONS:  true,
		sip.REFER:    true,
		sip.NOTIFY:   true,
	}

	var msgMethod sip.RequestMethod
	switch m := msg.(type) {
	case sip.Request:
		msgMethod = m.Method()
	case sip.Response:
		if cseq, ok := m.CSeq(); ok && !m.IsProvisional() {
			msgMethod = cseq.MethodName
		}
	}
	if len(msgMethod) > 0 {
		if _, ok := autoAppendMethods[msgMethod]; ok {
			hdrs := msg.GetHeaders("Allow")
			if len(hdrs) == 0 {
				allow := make(sip.AllowHeader, 0)
				for _, method := range srv.getAllowedMethods() {
					allow = append(allow, method)
				}

				if len(allow) > 0 {
					msg.AppendHeader(allow)
				}
			}

			hdrs = msg.GetHeaders("Supported")
			if len(hdrs) == 0 && len(srv.extensions) > 0 {
				msg.AppendHeader(&sip.SupportedHeader{
					Options: srv.extensions,
				})
			}
		}
	}

	if hdrs := msg.GetHeaders("User-Agent"); len(hdrs) == 0 {
		userAgent := sip.UserAgentHeader(srv.userAgent)
		msg.AppendHeader(&userAgent)
	}

	if hdrs := msg.GetHeaders("Content-Length"); len(hdrs) == 0 {
		msg.SetBody(msg.Body(), true)
	}
}

func (srv *ServerOpt) getAllowedMethods() []sip.RequestMethod {
	methods := []sip.RequestMethod{
		sip.INVITE,
		sip.ACK,
		sip.CANCEL,
	}
	added := map[sip.RequestMethod]bool{
		sip.INVITE: true,
		sip.ACK:    true,
		sip.CANCEL: true,
	}

	srv.hmu.RLock()
	for method := range srv.requestHandlers {
		if _, ok := added[method]; !ok {
			methods = append(methods, method)
		}
	}
	srv.hmu.RUnlock()

	return methods
}

type sipTransport struct {
	tpl transport.Layer
	srv *ServerOpt
}

func (tp *sipTransport) Messages() <-chan sip.Message {
	return tp.tpl.Messages()
}

func (tp *sipTransport) Send(msg sip.Message) error {
	return tp.srv.Send(msg)
}

func (tp *sipTransport) IsReliable(network string) bool {
	return tp.tpl.IsReliable(network)
}

func (tp *sipTransport) IsStreamed(network string) bool {
	return tp.tpl.IsStreamed(network)
}
