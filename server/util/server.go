package server

import (
	"encoding/binary"
	"github.com/leviathan1995/spleen/service"
	"log"
	"net"
	"strconv"
)

type ConnectionPool map[uint64]chan *net.TCPConn

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
	connectionPool ConnectionPool
}

func (pool ConnectionPool) Has(id uint64) bool {
	_, ok := pool[id]
	return ok
}

func NewServer(listenIP string, listenPort int, Rules []rule) *server {
	return &server{
		&service.Service{
			IP:   listenIP,
			Port: listenPort,
		},
		Rules,
		make(map[uint64]chan *net.TCPConn),
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

		/* The proxy should get the magic number and ID of the client first. */
		transBuf := make([]byte, service.IDBuf)
		err = s.TCPRead(conn, transBuf, service.IDBuf)
		if err != nil {
			_ = conn.Close()
			log.Println("Try to read the destination port failed.")
			continue
		}

		id := binary.LittleEndian.Uint64(transBuf)
		if ConnectionPool.Has(s.connectionPool, id) == false {
			_ = conn.Close()
			continue
		} else {
			s.connectionPool[id] <- conn
		}
	}
}

func (s *server) Listen() {
	for _, rule := range s.rules {
		tcpAddr, _ := net.ResolveTCPAddr("tcp", s.IP+":"+strconv.Itoa(rule.LocalPort))

		if ConnectionPool.Has(s.connectionPool, rule.ClientID) == false {
			s.connectionPool[rule.ClientID] = make(chan *net.TCPConn, 256)
		}

		go s.ListenForClient(tcpAddr, rule.ClientID, rule.MappingPort)
	}

	tcpAddr, _ := net.ResolveTCPAddr("tcp", s.IP+":"+strconv.Itoa(s.Port))
	s.ListenForIntranet(tcpAddr)
}

func (s *server) handleConn(cliConn *net.TCPConn, clientID uint64, transferPort uint64) {
	defer func(cliConn *net.TCPConn) {
		err := cliConn.Close()
		if err != nil {
			return
		}
	}(cliConn)

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
				log.Printf("Make a successful connection between the user [%s] and the intranet server[Client ID: %d - Port: %d].",
					cliConn.RemoteAddr().String(), clientID, transferPort)
				/* Transfer network packets. */
				go func() {
					errTransfer := s.TransferToTCP(cliConn, intranetConn, 0)
					if errTransfer != nil {
						_ = cliConn.Close()
						_ = intranetConn.Close()
						return
					}
				}()
				_ = s.TransferToTCP(intranetConn, cliConn, 0)
				return
			}
		}
	}
}
