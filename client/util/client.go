package client

import (
	"sync/atomic"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/leviathan1995/spleen/service"
)

type client struct {
	*service.Service
	clientID  int
	srvAddr   *net.TCPAddr
	limitRate map[int64]int64
}

func NewClient(clientID int, serverIP string, serverPort int, limitRate []string) *client {
	srvAddr, _ := net.ResolveTCPAddr("tcp", serverIP+":"+strconv.Itoa(serverPort))
	limits := make(map[int64]int64)
	for _, rates := range limitRate {
		rate := strings.Split(rates, ":")
		port, _ := strconv.Atoi(rate[0])
		speed, _ := strconv.Atoi(rate[1])
		limits[int64(port)] = int64(speed)
	}

	return &client{
		&service.Service{},
		clientID,
		srvAddr,
		limits,
	}
}

var connections int64 = 0;

func (c *client) Run() {
	log.Printf("Begin to running the client[%d]", c.clientID)
	for port, rate := range c.limitRate {
		log.Printf("The limiting rate configurations  Port: %d, Speed: %d KB/s", port, rate)
	}

	for {
		if atomic.LoadInt64(&connections) < 10 {
			for i := atomic.LoadInt64(&connections); i < 10; i++ {
				srvConn, err := c.DialSrv()
				if err != nil {
					log.Printf("Connect to the server %s:%d failed: %s. \n", c.srvAddr.IP.String(), c.srvAddr.Port, err)
					continue
				}
				log.Printf("Connect to the server %s:%d successful.\n", c.srvAddr.IP.String(), c.srvAddr.Port)
				srvConn.SetKeepAlive(true)
				srvConn.SetLinger(0)
				connections = atomic.AddInt64(&connections, 1)
				go c.handleConn(srvConn)
			}
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *client) DialSrv() (*net.TCPConn, error) {
	return net.DialTCP("tcp", nil, c.srvAddr)
}

func (c *client) handleConn(srvConn *net.TCPConn) {
	defer srvConn.Close()

	transBuf := make([]byte, 8)
	/* Send the ID of client to proxy. */
	binary.LittleEndian.PutUint64(transBuf, uint64(c.clientID))
	err := c.TCPWrite(srvConn, transBuf)
	if err != nil {
		connections = atomic.AddInt64(&connections, -1)
		log.Println("Try to send the ID of client to the proxy failed.")
		return
	}

	/* Waiting for the transfer port from proxy. */
	nRead, err := srvConn.Read(transBuf)
	connections = atomic.AddInt64(&connections, -1)
	if err != nil {
		log.Println("Try to read the destination port failed.")
		return
	}
	port := int64(binary.LittleEndian.Uint64(transBuf[:nRead]))

	/* Try to direct connect to the destination sever. */
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

	dstConn.SetKeepAlive(true)
	_ = dstConn.SetLinger(0)

	var limitRate int64

	if rate, found := c.limitRate[port]; found {
		limitRate = rate * 1024 /* bytes */
	}

	go func() {
		errTransfer := c.TransferToTCP(dstConn, srvConn, limitRate)
		if errTransfer != nil {
			return
		}
	}()
	err = c.TransferToTCP(srvConn, dstConn, 0)
}
