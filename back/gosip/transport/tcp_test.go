package transport_test

import (
	"fmt"
	"net"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ghettovoice/gosip/sip"
	"github.com/ghettovoice/gosip/testutils"
	"github.com/ghettovoice/gosip/timing"
	"github.com/ghettovoice/gosip/transport"
)

var _ = Describe("TcpProtocol", func() {
	var (
		output                    chan sip.Message
		errs                      chan error
		cancel                    chan struct{}
		protocol                  transport.Protocol
		client1, client2, client3 net.Conn
		wg                        *sync.WaitGroup
	)

	network := "tcp"
	port1 := 9060
	port2 := port1 + 1
	localTarget1 := transport.NewTarget(transport.DefaultHost, port1)
	localTarget2 := transport.NewTarget(transport.DefaultHost, port2)
	msg1 := "INVITE sip:bob@far-far-away.com SIP/2.0\r\n" +
		"Via: SIP/2.0/TCP pc33.far-far-away.com;branch=z9hG4bK776asdhds\r\n" +
		"To: \"Bob\" <sip:bob@far-far-away.com>\r\n" +
		"From: \"Alice\" <sip:alice@wonderland.com>;tag=1928301774\r\n" +
		"Content-Length: 12\r\n" +
		"\r\n" +
		"Hello world!"
	expectedMsg1 := "INVITE sip:bob@far-far-away.com SIP/2.0\r\n" +
		"Via: SIP/2.0/TCP pc33.far-far-away.com;branch=z9hG4bK776asdhds;received=%s\r\n" +
		"To: \"Bob\" <sip:bob@far-far-away.com>\r\n" +
		"From: \"Alice\" <sip:alice@wonderland.com>;tag=1928301774\r\n" +
		"Content-Length: 12\r\n" +
		"\r\n" +
		"Hello world!"
	msg2 := "BYE sip:bob@far-far-away.com SIP/2.0\r\n" +
		"Via: SIP/2.0/TCP pc33.far-far-away.com;branch=z9hG4bK776asdhds\r\n" +
		"To: \"Alice\" <sip:bob@far-far-away.com>\r\n" +
		"From: \"Bob\" <sip:alice@wonderland.com>;tag=1928301774\r\n" +
		"Content-Length: 4\r\n" +
		"\r\n" +
		"Bye!"
	expectedMsg2 := "BYE sip:bob@far-far-away.com SIP/2.0\r\n" +
		"Via: SIP/2.0/TCP pc33.far-far-away.com;branch=z9hG4bK776asdhds;received=%s\r\n" +
		"To: \"Alice\" <sip:bob@far-far-away.com>\r\n" +
		"From: \"Bob\" <sip:alice@wonderland.com>;tag=1928301774\r\n" +
		"Content-Length: 4\r\n" +
		"\r\n" +
		"Bye!"
	msg3 := "SIP/2.0 200 OK\r\n" +
		"Via: SIP/2.0/TCP pc33.far-far-away.com;branch=z9hG4bK776asdhds\r\n" +
		"CSeq: 2 INVITE\r\n" +
		"Call-ID: cheesecake1729\r\n" +
		"Max-Forwards: 65\r\n" +
		"Content-Length: 0\r\n" +
		"\r\n"
	broken := "BROKEN from hell.com SIP/2.0\r\n" +
		"Via: HELL\r\n" +
		"\r\n" +
		"THIS MESSAGE FROM HELL!\r\n"
	bullshit := "This is bullshit!\r\n"

	logger := testutils.NewLogrusLogger()

	timing.MockMode = true

	closeClients := func() {
		if client1 != nil {
			client1.Close()
		}
		if client2 != nil {
			client2.Close()
		}
		if client3 != nil {
			client3.Close()
		}
	}

	BeforeEach(func() {
		wg = new(sync.WaitGroup)
		output = make(chan sip.Message)
		errs = make(chan error)
		cancel = make(chan struct{})
		protocol = transport.NewTcpProtocol(output, errs, cancel, nil, logger)
	})
	AfterEach(func(done Done) {
		wg.Wait()
		select {
		case <-cancel:
		default:
			close(cancel)
		}
		<-protocol.Done()
		closeClients()
		close(output)
		close(errs)
		close(done)
	}, 3)

	Context("just initialized", func() {
		It("should has Network = TCP", func() {
			Expect(protocol.Network()).To(Equal("TCP"))
		})
		It("should be reliable", func() {
			Expect(protocol.Reliable()).To(BeTrue())
		})
		It("should be streamed", func() {
			Expect(protocol.Streamed()).To(BeTrue())
		})
	})

	Context(fmt.Sprintf("listens 2 target: %s, %s", localTarget1, localTarget2), func() {
		BeforeEach(func() {
			Expect(protocol.Listen(localTarget1)).To(Succeed())
			Expect(protocol.Listen(localTarget2)).To(Succeed())
			time.Sleep(time.Millisecond)
		})

		Context("when 3 clients connects and sends data", func() {
			BeforeEach(func() {
				client1 = testutils.CreateClient(network, localTarget1.Addr(), "")
				client2 = testutils.CreateClient(network, localTarget2.Addr(), "")
				client3 = testutils.CreateClient(network, localTarget1.Addr(), "")
				wg.Add(3)
				go func() {
					defer wg.Done()
					time.Sleep(time.Millisecond)
					testutils.WriteToConn(client1, []byte(msg1))
				}()
				go func() {
					defer wg.Done()
					time.Sleep(100 * time.Millisecond)
					testutils.WriteToConn(client2, []byte(msg2))
					time.Sleep(200 * time.Millisecond)
					testutils.WriteToConn(client2, []byte(bullshit))
					time.Sleep(200 * time.Millisecond)
					testutils.WriteToConn(client2, []byte(msg2))
				}()
				go func() {
					defer wg.Done()
					time.Sleep(50 * time.Millisecond)
					testutils.WriteToConn(client3, []byte(broken))
					time.Sleep(100 * time.Millisecond)
					testutils.WriteToConn(client3, []byte(msg3))
				}()
			})
			It("should pipe incoming messages and errors", func(done Done) {
				By(fmt.Sprintf("msg1 arrives on output from client1 %s -> %s", client1.LocalAddr().String(), localTarget1.Addr()))
				testutils.AssertMessageArrived(output, fmt.Sprintf(expectedMsg1, client1.LocalAddr().(*net.TCPAddr).IP), client1.LocalAddr().String(), localTarget1.Addr())
				By(fmt.Sprintf("broken message arrives from client3 and ignored %s -> %s", client3.LocalAddr().String(), localTarget1.Addr()))
				time.Sleep(time.Millisecond)
				By(fmt.Sprintf("msg2 arrives on output from client2 %s -> %s", client2.LocalAddr().String(), localTarget2.Addr()))
				testutils.AssertMessageArrived(output, fmt.Sprintf(expectedMsg2, client1.LocalAddr().(*net.TCPAddr).IP), client2.LocalAddr().String(), localTarget2.Addr())
				By(fmt.Sprintf("msg3 arrives on output from client3 %s -> %s", client3.LocalAddr().String(), localTarget1.Addr()))
				testutils.AssertMessageArrived(output, msg3, client3.LocalAddr().String(), localTarget1.Addr())
				By(fmt.Sprintf("bullshit arrives from client2 and ignored %s -> %s", client2.LocalAddr().String(), localTarget2.Addr()))
				time.Sleep(time.Millisecond)
				By(fmt.Sprintf("msg2 arrives on output from client2 %s -> %s", client2.LocalAddr().String(), localTarget2.Addr()))
				testutils.AssertMessageArrived(output, fmt.Sprintf(expectedMsg2, client1.LocalAddr().(*net.TCPAddr).IP), client2.LocalAddr().String(), localTarget2.Addr())
				// for i := 0; i < 4; i++ {
				//	select {
				//	case msg := <-output:
				//		fmt.Printf("\n-------------------------------\n%s\n-------------------------------------\n\n", msg)
				//	case err := <-errs:
				//		fmt.Printf("\n-------------------------------\n%s\n-------------------------------------\n\n", err)
				//	}
				// }
				close(done)
			}, 3)
		})

		Context("when client1 sends invite request", func() {
			BeforeEach(func() {
				client1 = testutils.CreateClient(network, localTarget1.Addr(), "")
				wg.Add(1)
				go func() {
					defer wg.Done()
					time.Sleep(100 * time.Millisecond)
					testutils.WriteToConn(client1, []byte(msg1))
				}()
			})
			It("should receive message and response with 200 OK", func(done Done) {
				By("msg1 arrives")
				testutils.AssertMessageArrived(output, fmt.Sprintf(expectedMsg1, client1.LocalAddr().(*net.TCPAddr).IP), client1.LocalAddr().String(), localTarget1.Addr())

				By("prepare response 200 OK")
				clientTarget, err := transport.NewTargetFromAddr(client1.LocalAddr().String())
				Expect(clientTarget).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				msg := sip.NewResponse(
					"",
					"SIP/2.0",
					200,
					"OK",
					[]sip.Header{
						&sip.CSeq{SeqNo: 2, MethodName: sip.INVITE},
					},
					"",
					nil,
				)
				twg := new(sync.WaitGroup)
				twg.Add(2)
				go func() {
					defer twg.Done()
					By("sends response 200 OK")
					Expect(protocol.Send(clientTarget, msg)).To(Succeed())
				}()
				go func() {
					defer twg.Done()
					buf := make([]byte, 65535)
					By("client server waiting 200 OK")
					for {
						num, err := client1.Read(buf)
						Expect(err).ToNot(HaveOccurred())
						Expect(num).To(Equal(len(msg.String())))
						data := append([]byte{}, buf[:num]...)
						Expect(string(data)).To(Equal(msg.String()))
						return
					}
				}()
				twg.Wait()
				close(done)
			}, 3)
		})

		Context("after cancel signal received", func() {
			BeforeEach(func() {
				time.Sleep(time.Millisecond)
				close(cancel)
			})
			It("should resolve Done chan", func(done Done) {
				<-protocol.Done()
				close(done)
			})
		})
	})
})
