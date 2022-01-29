package server

import (
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/leviathan1995/spleen/service"
)

type ConnectionPool map[int64]chan *net.TCPConn

type server struct {
	*service.Service
	mappingPort    []string
	connectionPool ConnectionPool
}

func (pool ConnectionPool) Has(id int64) bool {
	_, ok := pool[id]
	return ok
}

func NewServer(listenIP string, listenPort int, mappingPort []string) *server {
	return &server{
		&service.Service{
			IP:   listenIP,
			Port: listenPort,
		},
		mappingPort,
		make(map[int64]chan *net.TCPConn),
	}
}

func (s *server) ListenForClient(tcpAddr *net.TCPAddr, clientID, transferPort string) {
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Printf("The server try to listening at %s:%d failed.", tcpAddr.IP.String(), tcpAddr.Port)
		return
	} else {
		log.Printf("The server listening at %s:%d for [Client ID: %s - Port: %s] successful.", tcpAddr.IP.String(), tcpAddr.Port, clientID, transferPort)
	}
	defer listener.Close()

	for {
		cliConn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}
		port, _ := strconv.Atoi(transferPort)
		id, _ := strconv.Atoi(clientID)
		go s.handleConn(cliConn, int64(id), uint64(port))
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

		id := int64(binary.LittleEndian.Uint64(transBuf))
		if ConnectionPool.Has(s.connectionPool, id) == false {
			_ = conn.Close()
			continue
		} else {
			s.connectionPool[id] <- conn
		}
	}
}

func (s *server) Listen() {
	for _, ports := range s.mappingPort {
		p := strings.Split(ports, ":")
		tcpAddr, _ := net.ResolveTCPAddr("tcp", s.IP+":"+p[0])

		id, _ := strconv.Atoi(p[1])
		if ConnectionPool.Has(s.connectionPool, int64(id)) == false {
			s.connectionPool[int64(id)] = make(chan *net.TCPConn, 512)
		}

		go s.ListenForClient(tcpAddr, p[1], p[2])
	}

	tcpAddr, _ := net.ResolveTCPAddr("tcp", s.IP+":"+strconv.Itoa(s.Port))
	s.ListenForIntranet(tcpAddr)
}

func (s *server) handleConn(cliConn *net.TCPConn, clientID int64, transferPort uint64) {
	defer cliConn.Close()

	for {
		select {
		case intranetConn := <-s.connectionPool[clientID]:
			_ = intranetConn.SetLinger(0)

			/* Send the transfer port to intranet server . */
			portBuf := make([]byte, service.PortBuf)
			binary.LittleEndian.PutUint64(portBuf, transferPort)
			err := s.TCPWrite(intranetConn, portBuf)
			if err != nil {
				_ = intranetConn.Close()
				/* Waiting for the connection again. */
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
