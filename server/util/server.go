package server

import (
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/leviathan1995/spleen/service"
)

var connectionPool = make(chan *net.TCPConn, 512)

type server struct {
	*service.Service
	mappingPort []string
}

func NewServer(listenIP string, listenPort int, mappingPort []string) *server {
	return &server{
		&service.Service{
			IP:   listenIP,
			Port: listenPort,
		},
		mappingPort,
	}
}

func (s *server) ListenOnPort(tcpAddr *net.TCPAddr, transferPort string) {
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Printf("The server try to listening at %s:%d failed.", tcpAddr.IP.String(), tcpAddr.Port)
		return
	} else {
		log.Printf("The server listening at %s:%d successful.", tcpAddr.IP.String(), tcpAddr.Port)
	}
	defer listener.Close()

	for {
		cliConn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}
		port, _ := strconv.Atoi(transferPort)
		go s.handleConn(cliConn, uint64(port))
	}
}

func (s *server) ListenForIntranet(tcpAddr *net.TCPAddr) {
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Printf("The server try to listening for the intranet server at %s:%d failed.", s.IP, s.Port)
		return
	} else {
		log.Printf("The server listening for the intranet server at %s:%d successful.", s.IP, s.Port)
	}
	defer listener.Close()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}
		connectionPool <- conn
	}
}

func (s *server) Listen() {
	for _, ports := range s.mappingPort {
		p := strings.Split(ports, ":")
		tcpAddr, _ := net.ResolveTCPAddr("tcp", s.IP+":"+p[0])
		go s.ListenOnPort(tcpAddr, p[1])
	}

	tcpAddr, _ := net.ResolveTCPAddr("tcp", s.IP+":"+strconv.Itoa(s.Port))
	s.ListenForIntranet(tcpAddr)
}

func (s *server) handleConn(cliConn *net.TCPConn, transferPort uint64) {
	defer cliConn.Close()

	for i := 0; i < 20; i++ {
		intranetConn := <-connectionPool
		if intranetConn != nil {
			_ = intranetConn.SetLinger(0)

			/* Send the mapping port to intranet server . */
			log.Printf("Send the mapping port %d to intranet server.\n", transferPort)
			portBuf := make([]byte, 8)
			binary.LittleEndian.PutUint64(portBuf, transferPort)
			err := s.TCPWrite(intranetConn, portBuf)
			if err != nil {
				intranetConn.Close()
				continue
			}

			log.Print("Make a successful connection between client and the intranet server.")
			/* Transfer network packets. */
			go func() {
				errTransfer := s.TransferToTCP(cliConn, intranetConn)
				if errTransfer != nil {
					intranetConn.Close()
					return
				}
			}()
			err = s.TransferToTCP(intranetConn, cliConn)
		}
	}
}
