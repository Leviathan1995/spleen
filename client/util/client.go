package client

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/leviathan1995/spleen/service"
)

type client struct {
	*service.Service
	srvAddr *net.TCPAddr
}

func NewClient(serverIP string, serverPort int) *client {
	srvAddr, _ := net.ResolveTCPAddr("tcp", serverIP+":"+strconv.Itoa(serverPort))
	return &client{
		&service.Service{},
		srvAddr,
	}
}

func (c *client) Run() {
	for {
		if len(connectionPool) < 10 {
			srvConn, err := c.DialSrv()
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("Connect to the server %s:%d successful.\n", c.srvAddr.IP.String(), c.srvAddr.Port)
			connectionPool <- srvConn
			go c.handleConn(srvConn)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

var connectionPool = make(chan net.Conn, 512)

func (c *client) DialSrv() (*net.TCPConn, error) {
	return net.DialTCP("tcp", nil, c.srvAddr)
}

func (c *client) handleConn(srvConn *net.TCPConn) {
	defer srvConn.Close()

	/* Get the transfer port. */
	portBuf := make([]byte, 32)
	nRead, err := srvConn.Read(portBuf)
	if err != nil {
		return
	}
	_ = <-connectionPool /* Remove a connection from pool. */

	/* Try to direct connect to the destination sever. */
	dstAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1"+":"+string(portBuf[:nRead]))
	log.Printf("Try to connect %s:%d.\n", dstAddr.IP.String(), dstAddr.Port)
	dstConn, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		log.Printf("Connect to %s:%d failed.", dstAddr.IP.String(), dstAddr.Port)
		return
	} else {
		log.Printf("Connect to the destination address %s:%d successful.", dstAddr.IP, dstAddr.Port)
	}

	defer dstConn.Close()
	_ = dstConn.SetLinger(0)

	go func() {
		errTransfer := c.TransferToTCP(dstConn, srvConn)
		if errTransfer != nil {
			log.Println(errTransfer.Error())
		}
	}()
	err = c.TransferToTCP(srvConn, dstConn)
}
