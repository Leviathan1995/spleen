package service

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
)

const BUFFERSIZE = 1024 * 4

type Service struct {
	IP   string
	Port int
}

func (s *Service) ParseSOCKS5(cliConn net.Conn) (*net.TCPAddr, error) {
	buf := make([]byte, BUFFERSIZE)

	readCount, errRead := cliConn.Read(buf)
	if errRead == io.EOF {
		return &net.TCPAddr{}, errRead
	}
	if readCount > 0 && errRead == nil {
		if buf[0] != 0x05 {
			/* Version Number */
			return &net.TCPAddr{}, errors.New("Only Support SOCKS5.")
		} else {
			/* [SOCKS5, NO AUTHENTICATION REQUIRED]  */
			_, errWrite := cliConn.Write([]byte{0x05, 0x00})
			if errWrite != nil {
				return &net.TCPAddr{}, errors.New("Response SOCKS5 failed at the first stage.")
			}
		}
	}

	readCount, errRead = cliConn.Read(buf)
	if errRead == io.EOF {
		return &net.TCPAddr{}, errRead
	}
	if readCount > 0 && errRead == nil {
		if buf[1] != 0x01 {
			/* Only support CONNECT method */
			return &net.TCPAddr{}, errors.New("Only support CONNECT method.")
		}

		var dstIP []byte
		switch buf[3] { /* checking ATYPE */
		case 0x01: /* IPv4 */
			dstIP = buf[4 : 4+net.IPv4len]
		case 0x03: /* DOMAINNAME */
			ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:readCount-2]))
			if err != nil {
				return &net.TCPAddr{}, errors.New("Parse IP from DomainName failed.")
			}
			dstIP = ipAddr.IP
		case 0x04: /* IPV6 */
			dstIP = buf[4 : 4+net.IPv6len]
		default:
			return &net.TCPAddr{}, errors.New("Wrong DST.ADDR and DST.PORT")
		}
		dstPort := buf[readCount-2 : readCount]

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
		readCount, err := cliConn.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			} else {
				return nil
			}
		}
		if readCount > 0 {
			_, err := dstConn.Write(buf[0:readCount])
			if err != nil {
				return err
			}
		}
	}
}

func (s *Service) TransferToTLS(dstConn *net.TCPConn, srcConn net.Conn) error {
	buf := make([]byte, BUFFERSIZE)
	for {
		readCount, err := dstConn.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			} else {
				return nil
			}
		}
		if readCount > 0 {
			_, err = srcConn.Write(buf[0:readCount])
			if err != nil {
				return err
			}
		}
	}
}