package service

import (
	"net"
	"time"
)

const TransferBuf = 1024 * 16
const PortBuf = 8
const IDBuf = 8

type Service struct {
	IP   string
	Port int
}

func (s *Service) TCPWrite(conn *net.TCPConn, buf []byte) error {
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

func (s *Service) TCPRead(conn *net.TCPConn, buf []byte, len int) error {
	nRead := 0
	for nRead < len {
		n, errRead := conn.Read(buf[nRead:])
		if errRead != nil {
			return errRead
		}
		nRead += n
	}
	return nil
}

func (s *Service) TransferToTCP(cliConn net.Conn, dstConn *net.TCPConn, limitRate uint64) error {
	buf := make([]byte, TransferBuf)
	var totalRead uint64
	var lastTime int64

	for {
		nRead, errRead := cliConn.Read(buf)
		if errRead != nil {
			return errRead
		}
		if nRead > 0 {
			errWrite := s.TCPWrite(dstConn, buf[0:nRead])
			if errWrite != nil {
				return errWrite
			}
			if limitRate > 0 {
				if totalRead > limitRate && ((time.Now().UnixNano()/1e6)-lastTime) >= 1000 {
					/* Reset the timeout. */
					totalRead = 0
					lastTime = time.Now().UnixNano() / 1e6 /* The millisecond */
				} else if totalRead > limitRate && ((time.Now().UnixNano()/1e6)-lastTime) < 1000 {
					/* Try to limit the rate of network. */
					timeout := 1000 - ((time.Now().UnixNano() / 1e6) - lastTime)
					time.Sleep(time.Duration(timeout) * time.Millisecond)
					totalRead = 0
					lastTime = time.Now().UnixNano() / 1e6 /* The millisecond */
				} else {
					totalRead += uint64(nRead)
				}
			}
		}
	}
}
