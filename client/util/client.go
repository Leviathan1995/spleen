package client

import (
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/leviathan1995/spleen/service"
)

type client struct {
	*service.Service
	clientID  int
	srvAddr   *net.TCPAddr
	limitRate map[int64]uint64
}

func NewClient(clientID int, serverIP string, serverPort int, limitRate []string) *client {
	srvAddr, _ := net.ResolveTCPAddr("tcp", serverIP+":"+strconv.Itoa(serverPort))
	limits := make(map[int64]uint64)
	for _, rates := range limitRate {
		rate := strings.Split(rates, ":")
		port, _ := strconv.Atoi(rate[0])
		speed, _ := strconv.Atoi(rate[1])
		limits[int64(port)] = uint64(speed)
	}

	return &client{
		&service.Service{},
		clientID,
		srvAddr,
		limits,
	}
}

var freeConnections int64 = 0

func (c *client) Run() {
	log.Printf("Begin to running the client[%d]", c.clientID)
	for port, rate := range c.limitRate {
		log.Printf("The limiting rate configurations  Port: %d, Speed: %d KB/s", port, rate)
	}

	for {
		if atomic.LoadInt64(&freeConnections) < 10 {
			for i := atomic.LoadInt64(&freeConnections); i < 10; i++ {
				srvConn, err := c.DialSrv()
				if err != nil {
					log.Printf("Connect to the proxy %s:%d failed: %s. \n", c.srvAddr.IP.String(), c.srvAddr.Port, err)
					continue
				}
				log.Printf("Connect to the proxy %s:%d successful.\n", c.srvAddr.IP.String(), c.srvAddr.Port)
				_ = srvConn.SetLinger(0)
				_ = srvConn.SetKeepAlive(true)
				_ = srvConn.SetKeepAlivePeriod(2 * time.Second)
				atomic.AddInt64(&freeConnections, 1)
				go c.handleConn(srvConn)
			}
		} else {
			log.Printf("Currently, We still have %d active connections.", atomic.LoadInt64(&freeConnections))
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *client) DialSrv() (*net.TCPConn, error) {
	return net.DialTCP("tcp", nil, c.srvAddr)
}

func (c *client) handleConn(srvConn *net.TCPConn) {
	/* Send the ID of client to proxy. */
	transBuf := make([]byte, service.IDBuf)
	binary.LittleEndian.PutUint64(transBuf, uint64(c.clientID))
	err := c.TCPWrite(srvConn, transBuf)
	if err != nil {
		atomic.AddInt64(&freeConnections, -1)
		_ = srvConn.Close()
		log.Println("Try to send the ID of client to the proxy failed.")
		return
	}

	/* It has to wait 3600 seconds before get the transfer port from the proxy. */
	_ = srvConn.SetDeadline(time.Now().Add(3600 * time.Second))
	err = c.TCPRead(srvConn, transBuf, service.PortBuf)
	atomic.AddInt64(&freeConnections, -1)
	if err != nil {
		_ = srvConn.Close()
		log.Println("Try to read destination port from the proxy failed or maybe the connection is closed.")
		return
	}
	port := int64(binary.LittleEndian.Uint64(transBuf))
	/* Handshake successful, remove the deadline. */
	_ = srvConn.SetDeadline(time.Time{})

	/* Try to direct connect to the destination sever. */
	dstAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		_ = srvConn.Close()
		log.Printf("Try to resolve TCPAddr %s failed: %s.\n", "localhost"+":"+strconv.FormatInt(port, 10), err.Error())
		return
	}

	dstConn, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		_ = srvConn.Close()
		log.Printf("Connect to localhost:%d failed.", dstAddr.Port)
		return
	} else {
		log.Printf("Connect to the destination address localhost:%d successful.", dstAddr.Port)
	}

	_ = dstConn.SetLinger(0)
	_ = dstConn.SetKeepAlive(true)
	_ = dstConn.SetKeepAlivePeriod(60 * time.Second)

	var limitRate uint64
	if rate, found := c.limitRate[port]; found {
		limitRate = rate * 1024 /* bytes */
	}

	go func() {
		_ = c.TransferToTCP(dstConn, srvConn, limitRate)
	}()

	_ = c.TransferToTCP(srvConn, dstConn, 0)
	_ = srvConn.Close()
	_ = dstConn.Close()

	return
}
