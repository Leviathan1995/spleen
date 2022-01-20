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
		/* The proxy should get the ID of the client first. */
		transBuf := make([]byte, 8)
		nRead, err := conn.Read(transBuf)
		if err != nil {
			log.Println("Try to read the destination port failed.")
			return
		}
		id := int64(binary.LittleEndian.Uint64(transBuf[:nRead]))
		s.connectionPool[id] <- conn
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

	select {
	case intranetConn := <-s.connectionPool[clientID]:
		_ = intranetConn.SetLinger(0)

		/* Send the transfer port to intranet server . */
		portBuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(portBuf, transferPort)
		err := s.TCPWrite(intranetConn, portBuf)
		if err != nil {
			intranetConn.Close()
			for {
				/* Close all connections from this client. */
				select {
				case intranetConn = <-s.connectionPool[clientID]:
					intranetConn.Close()
				default:
					return
				}
			}
		}

		log.Printf("Make a successful connection between the user and the intranet server[Client ID: %d - Port: %d].", clientID, transferPort)
		/* Transfer network packets. */
		go func() {
			errTransfer := s.TransferToTCP(cliConn, intranetConn, 0)
			if errTransfer != nil {
				intranetConn.Close()
				return
			}
		}()
		err = s.TransferToTCP(intranetConn, cliConn, 0)
		return
	default:
		log.Printf("Currently, Do not have any active connection from the intranet server[Client ID: %d - Port: %d].", clientID, transferPort)
	}
}
