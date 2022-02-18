package server

import (
	"encoding/binary"
	"github.com/leviathan1995/spleen/service"
	"log"
	"net"
	"strconv"
	"time"
)

type rule struct {
	ClientID    uint64
	LocalPort   int
	MappingPort int
}

type Configuration struct {
	ServerIP   string
	ServerPort int
	Rules      []rule
}

type server struct {
	*service.Service
	rules          []rule
	connectionPool []chan *net.TCPConn
}

func NewServer(listenIP string, listenPort int, Rules []rule) *server {
	return &server{
		&service.Service{
			IP:   listenIP,
			Port: listenPort,
		},
		Rules,
		make([]chan *net.TCPConn, 1024),
	}
}

func (s *server) ListenForClient(tcpAddr *net.TCPAddr, clientID uint64, transferPort int) {
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Printf("The server try to listening at %s:%d failed.", tcpAddr.IP.String(), tcpAddr.Port)
		return
	} else {
		log.Printf("The server listening at %s:%d for [Client ID: %d - Port: %d] successful.", tcpAddr.IP.String(), tcpAddr.Port, clientID, transferPort)
	}
	defer listener.Close()

	s.connectionPool[clientID] = make(chan *net.TCPConn, 256)

	for {
		cliConn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}

		_ = cliConn.SetLinger(0)

		go s.handleConn(cliConn, clientID, uint64(transferPort))
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

		_ = conn.SetLinger(0)

		/* The proxy should get the ID of the client first. */
		transBuf := make([]byte, service.IDBuf)
		err = s.TCPRead(conn, transBuf, service.IDBuf)
		if err != nil {
			_ = conn.Close()
			log.Println("Try to read the destination port failed.")
			continue
		}

		clientID := binary.LittleEndian.Uint64(transBuf)
		if clientID < 1024 {
			s.connectionPool[clientID] <- conn
		} else {
			_ = conn.Close()
			continue
		}
	}
}

func (s *server) Listen() {
	for _, rule := range s.rules {
		tcpAddr, _ := net.ResolveTCPAddr("tcp", s.IP+":"+strconv.Itoa(rule.LocalPort))

		go s.ListenForClient(tcpAddr, rule.ClientID, rule.MappingPort)
	}

	tcpAddr, _ := net.ResolveTCPAddr("tcp", s.IP+":"+strconv.Itoa(s.Port))
	s.ListenForIntranet(tcpAddr)
}

func (s *server) handleConn(cliConn *net.TCPConn, clientID uint64, transferPort uint64) {
	for {
		select {
		case intranetConn := <-s.connectionPool[clientID]:
			_ = intranetConn.SetLinger(0)

			/* Send the transfer port to intranet server . */
			portBuf := make([]byte, service.PortBuf)
			binary.LittleEndian.PutUint64(portBuf, transferPort)
			err := s.TCPWrite(intranetConn, portBuf)
			if err != nil {
				/* Waiting for the new connection again. */
				_ = intranetConn.Close()
				continue
			} else {
				log.Printf("Make a successful connection between the user[%s] and the intranet server[Client ID: %d - Port: %d].",
					cliConn.RemoteAddr().String(), clientID, transferPort)
				/* Transfer network packets. */
				go func() {
					_ = s.TransferToTCP(cliConn, intranetConn, 0)
					_ = intranetConn.Close()
					_ = cliConn.Close()
				}()

				_ = s.TransferToTCP(intranetConn, cliConn, 0)
				return
			}
		case <-time.After(30 * time.Second):
			log.Printf("Currently, We don't have any active connection from the intranet server[Client ID: %d - Port: %d].",
				clientID, transferPort)
			_ = cliConn.Close()
		}
	}
}
