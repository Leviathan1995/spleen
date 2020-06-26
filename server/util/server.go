package server

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/spleen/service"
	"io/ioutil"
	"log"
	"net"
	"strconv"
)

type server struct {
	*service.Service
}

func NewServer(localIP string, localPort int) *server {
	return &server{
		&service.Service{
			IP:   localIP,
			Port: localPort,
		},
	}
}

func (s *server) Listen() error {
	log.Printf("Server local address: %s:%d", s.IP, s.Port)

	cert, err := tls.LoadX509KeyPair("/etc/server.pem", "/etc/server.key")
	if err != nil {
		log.Println(err)
		return err
	}

	certBytes, err := ioutil.ReadFile("/etc/client.pem")
	if err != nil {
		panic("Unable to read cert.pem")
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		panic("failed to parse root certificate")
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCertPool,
	}

	listener, err := tls.Listen("tcp", s.IP + ":" + strconv.Itoa(s.Port), config)
	if err != nil {
		return err
	} else {
		log.Printf("Server listen at %s:%d successed.", s.IP, s.Port)
	}
	defer listener.Close()

	for {
		cliConn, err := listener.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		}
		go s.handleConn(cliConn)
	}

}

func (s *server) handleConn(cliConn net.Conn) {
	defer cliConn.Close()

	dstAddr, errParse := s.ParseSOCKS5(cliConn)
	if errParse != nil {
		log.Printf("Parse SOCKS5 failed: %s", errParse.Error())
		return
	}

	/* Try to direct connect to the destination sever. */
	dstConn, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		log.Printf("Connect to %s:%d failed.", dstAddr.IP.String(), dstAddr.Port)
		return
	} else {
		log.Printf("Connect to the destination address success %s:%d.", dstAddr.IP, dstAddr.Port)
	}

	defer dstConn.Close()
	_ = dstConn.SetLinger(0)

	/* Connect to the destination sever success. */
	errWrite := s.TLSWrite(cliConn, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if errWrite != nil {
		log.Println("Server reply the SOCKS5 protocol failed at the second stage.")
		return
	}

	go func() {
		errTransfer := s.TransferToTCP(cliConn, dstConn)
		if errTransfer != nil {
			log.Println(errTransfer.Error())
		}
	}()
	err = s.TransferToTLS(dstConn, cliConn)
}
