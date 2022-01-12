package service

import (
	"io"
	"net"
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

func (s *Service) TransferToTCP(cliConn net.Conn, dstConn *net.TCPConn) error {
	buf := make([]byte, BUFFERSIZE)
	for {
		nRead, err := cliConn.Read(buf)
		if err != nil {
			return err
		}
		if nRead > 0 {
			errWrite := s.TCPWrite(dstConn, buf[0:nRead])
			if err != nil {
				if errWrite == io.EOF {
					return nil
				} else {
					return errWrite
				}
			}
		}
	}
}
