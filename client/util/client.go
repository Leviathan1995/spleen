package client

import (
	"encoding/binary"
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
			srvConn.SetKeepAlive(true)
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
	portBuf := make([]byte, 8)
	nRead, err := srvConn.Read(portBuf)
	_ = <-connectionPool /* Remove a connection from pool. */
	if err != nil {
		log.Println("Try to read the destination port failed.")
		return
	}
	port := int64(binary.LittleEndian.Uint64(portBuf[:nRead]))

	/* Try to direct connect to the destination sever. */
	log.Printf("Try to connect %s.\n", "localhost"+":"+strconv.Itoa(int(port)))
	dstAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		log.Printf("Try to resolve TCPAddr %s failed: %s.\n", "localhost"+":"+string(port), err.Error())
		return
	}

	dstConn, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		log.Printf("Connect to localhost:%d failed.", dstAddr.Port)
		return
	} else {
		log.Printf("Connect to the destination address localhost:%d successful.", dstAddr.Port)
	}

	defer dstConn.Close()
	_ = dstConn.SetLinger(0)

	go func() {
		errTransfer := c.TransferToTCP(dstConn, srvConn)
		if errTransfer != nil {
			return
		}
	}()
	err = c.TransferToTCP(srvConn, dstConn)
}
