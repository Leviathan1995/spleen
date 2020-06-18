package client

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/spleen/service"
	"io/ioutil"
	"log"
	"net"
	"strconv"
)

type client struct {
	*service.Service
	srvAddr *net.TCPAddr
}

func NewClient(localIP string, localPort int, serverIP string, serverPort int) *client {
	serAddr, _ := net.ResolveTCPAddr("tcp", serverIP + ":" + strconv.Itoa(serverPort))
	return &client{
		&service.Service{
			IP:   localIP,
			Port: localPort,
		},
		serAddr,
	}
}

func (c *client) Listen() error {
	log.Printf("Client local address: %s:%d", c.IP, c.Port)

	cert, err := tls.LoadX509KeyPair("client.pem", "client.key")
	if err != nil {
		log.Println(err)
		return err
	}
	certBytes, err := ioutil.ReadFile("client.pem")
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
		go c.handleConn(userConn, conf)
	}

}

func (c *client) handleConn(userConn *net.TCPConn, conf *tls.Config) {
	defer userConn.Close()

	srvConn, err := tls.Dial("tcp", c.srvAddr.String(), conf)
	if err != nil {
		log.Println(err)
		return
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


