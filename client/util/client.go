package client

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/spleen/service"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"
)

type client struct {
	*service.Service
	srvAddr *net.TCPAddr
	conf *tls.Config
}

func NewClient(localIP string, localPort int, serverIP string, serverPort int) *client {
	serAddr, _ := net.ResolveTCPAddr("tcp", serverIP + ":" + strconv.Itoa(serverPort))
	return &client{
		&service.Service{
			IP:   localIP,
			Port: localPort,
		},
		serAddr,
		nil,
	}
}

func (c *client) Listen() error {
	log.Printf("Client local address: %s:%d", c.IP, c.Port)

	cert, err := tls.LoadX509KeyPair("./client.pem", "./client.key")
	if err != nil {
		log.Println(err)
		return err
	}

	/* Parse .pem */
	certBytes, err := ioutil.ReadFile("./client.pem")
	if err != nil {
		panic("Unable to read cert.pem")
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		panic("failed to parse root certificate")
	}
	conf := &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	c.conf = conf

	/* Listen */
	cliAddr, _ := net.ResolveTCPAddr("tcp", c.IP + ":" + strconv.Itoa(c.Port))
	listener, err := net.ListenTCP("tcp", cliAddr)
	if err != nil {
		return err
	} else {
		log.Printf("Client listen at %s:%d successed.", c.IP, c.Port)
	}
	defer listener.Close()

	for {
		userConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err.Error())
			continue
		}
		go c.handleConn(userConn)
	}

}
var connectionPool = make(chan net.Conn, 10)

func init() {
	go func() {
		for range time.Tick(5 * time.Second) {
			p := <-connectionPool	/* Discard the idle connection */
			p.Close()
		}
	}()
}

func (c *client) DialSrv() (net.Conn, error) {
	return tls.Dial("tcp", c.srvAddr.String(), c.conf)
}

func (c *client) newSrvConn() (net.Conn, error) {
	if len(connectionPool) < 10 {
		go func() {
			for i := 0; i < 2; i++ {
				proxy, err := c.DialSrv()
				if err != nil {
					log.Println(err)
					return
				}
				connectionPool <- proxy
			}
		}()
	}

	select {
	case pc := <-connectionPool:
		return pc, nil
	case <-time.After(100 * time.Millisecond):
		return c.DialSrv()
	}
}

func (c *client) handleConn(userConn *net.TCPConn) {
	defer userConn.Close()

	srvConn, err := c.newSrvConn()
	if err != nil {
		log.Println(err)
		srvConn, err = c.newSrvConn()
		if err != nil {
			log.Println(err)
			return
		}
	}
	defer srvConn.Close()

	go func() {
		errTransfer := c.TransferToTLS(userConn, srvConn)
		if errTransfer != nil {
			log.Println(errTransfer.Error())
		}
	}()
	err = c.TransferToTCP(srvConn, userConn)
}


