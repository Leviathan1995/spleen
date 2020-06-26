package service

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
)

const BUFFERSIZE = 1024 * 8

type Service struct {
	IP   string
	Port int
}


func (s *Service) TLSWrite(conn net.Conn, buf []byte) (error) {
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

func (s *Service) TCPWrite(conn *net.TCPConn, buf []byte) (error) {
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

func (s *Service) ParseSOCKS5(cliConn net.Conn) (*net.TCPAddr, error) {
	buf := make([]byte, BUFFERSIZE)

	nRead, errRead := cliConn.Read(buf)
	if errRead != nil {
		return &net.TCPAddr{}, errors.New("Read SOCKS5 failed at the first stage.")
	}
	if nRead > 0 {
		if buf[0] != 0x05 {
			/* Version Number */
			return &net.TCPAddr{}, errors.New("Only support SOCKS5 protocol.")
		} else {
			/* [SOCKS5, NO AUTHENTICATION REQUIRED]  */
			errWrite := s.TLSWrite(cliConn, []byte{0x05, 0x00})
			if errWrite != nil {
				return &net.TCPAddr{}, errors.New("Response SOCKS5 failed at the first stage.")
			}
		}
	}

	nRead, errRead = cliConn.Read(buf)
	if errRead != nil {
		return &net.TCPAddr{}, errors.New("Read SOCKS5 failed at the second stage.")
	}
	if nRead > 0 {
		if buf[1] != 0x01 {
			/* Only support CONNECT method */
			return &net.TCPAddr{}, errors.New("Only support CONNECT method.")
		}

		var dstIP []byte
		switch buf[3] { /* checking ATYPE */
		case 0x01: /* IPv4 */
			dstIP = buf[4 : 4+net.IPv4len]
		case 0x03: /* DOMAINNAME */
			ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:nRead-2]))
			if err != nil {
				return &net.TCPAddr{}, errors.New("Parse IP from DomainName failed.")
			}
			dstIP = ipAddr.IP
		case 0x04: /* IPV6 */
			dstIP = buf[4 : 4+net.IPv6len]
		default:
			return &net.TCPAddr{}, errors.New("Wrong DST.ADDR and DST.PORT")
		}
		dstPort := buf[nRead-2 : nRead]

		if buf[1] == 0x01 {
			/* TCP over SOCKS5 */
			dstAddr := &net.TCPAddr{
				IP:   dstIP,
				Port: int(binary.BigEndian.Uint16(dstPort)),
			}
			return dstAddr, errRead
		} else {
			log.Println("Only support CONNECT method.")
			return &net.TCPAddr{}, errRead
		}
	}
	return &net.TCPAddr{}, errRead
}

func (s *Service) TransferToTCP(cliConn net.Conn, dstConn *net.TCPConn) error {
	buf := make([]byte, BUFFERSIZE)
	for {
		nRead, err := cliConn.Read(buf)
		if err != nil {
			return err
		}
		if nRead > 0 {
			errWrite := s.TCPWrite(dstConn, buf[0 : nRead])
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

func (s *Service) TransferToTLS(dstConn *net.TCPConn, srcConn net.Conn) error {
	buf := make([]byte, BUFFERSIZE)
	for {
		nRead, errRead := dstConn.Read(buf)
		if errRead != nil {
			if errRead == io.EOF {
				return nil
			} else {
				return errRead
			}
		}
		if nRead > 0 {
			errWrite := s.TLSWrite(srcConn, buf[0 : nRead])
			if errWrite != nil {
				if errWrite == io.EOF {
					return nil
				} else {
					return errWrite
				}
			}
		}
	}
}