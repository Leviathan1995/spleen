package main

import (
	"encoding/binary"
	"math/rand"
	"net"
	"sync"
	"time"
)

var wg sync.WaitGroup
var addr = "127.0.0.1:5003"

func dialSrv() (*net.TCPConn, error) {
	srvAddr, _ := net.ResolveTCPAddr("tcp", addr)
	return net.DialTCP("tcp", nil, srvAddr)
}

func tcpWrite(conn *net.TCPConn, buf []byte) error {
	nWrite := 0
	nBuffer := len(buf)
	for nWrite < nBuffer {
		n, errWrite := conn.Write(buf[nWrite:])
		if errWrite != nil {
			return errWrite
		}
		nWrite += n
	}
	return nil
}

func dial() {
	defer wg.Done()
	for i := 0; i <= 100000; i++ {
		srvConn, err := dialSrv()

		if err != nil {
      continue
    } else {
			/* Send the ID of client to proxy. */
			transBuf := make([]byte, 8)
			binary.LittleEndian.PutUint64(transBuf, uint64(rand.Intn(10)))
			_ = tcpWrite(srvConn, transBuf)
			time.Sleep(1 * time.Second)
			_ = srvConn.Close()
		}
	}
}

func main() {
	for i := 0; i <= 3; i++ {
		wg.Add(1)
		time.Sleep(1 * time.Second)
		go dial()
	}
	wg.Wait()
}
