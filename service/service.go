package service

import (
	"net"
	"time"
)

const BUFFERSIZE = 1024 * 16

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

func (s *Service) TransferToTCP(cliConn net.Conn, dstConn *net.TCPConn, limitRate int64) error {
	var totalRead, lastTime int64
	buf := make([]byte, BUFFERSIZE)

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
			if limitRate != 0 {
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
					totalRead += int64(nRead)
				}
			}
		}
	}
}
