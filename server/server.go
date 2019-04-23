package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/tywkeene/http-rfc/connection"
)

type HTTPVersion struct {
	Major int
	Minor int
}

type Server struct {
	Version        *HTTPVersion
	Listener       *net.TCPListener
	BindAddr       *net.TCPAddr
	ConnPool       *connection.ConnectionPool
	ConnErr        chan error
	ReadBufferSize uint32
}

func ParseHTTPVersion(major, minor int) *HTTPVersion {
	return &HTTPVersion{
		Major: major,
		Minor: minor,
	}
}

func (v *HTTPVersion) String() string {
	return "HTTP/" + strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor)
}

func (v *HTTPVersion) GetMajor() int {
	return v.Major
}

func (v *HTTPVersion) GetMinor() int {
	return v.Major
}

func NewServer(bindAddr string) *Server {

	values := strings.Split(bindAddr, ":")

	ip := net.ParseIP(values[0])
	port, _ := strconv.Atoi(values[1])

	addr := &net.TCPAddr{
		IP:   ip,
		Port: port,
	}

	return &Server{
		Version:  ParseHTTPVersion(2, 0),
		BindAddr: addr,
	}
}

func (s *Server) HandleConnError() {
	for {
		select {
		case err := <-s.ConnErr:
			log.Printf("Error handing connection: %q\n", err.Error())
		}
	}
}

func (s *Server) HandleConnection(rawConn *net.TCPConn) {

	c := s.ConnPool.AddConnection(rawConn)

	_, err := c.ReadRequest()
	if err != nil {
		s.ConnErr <- err
		c.Close()
		return
	}
}

func (s *Server) ServeHTTP() error {
	var err error

	s.Listener, err = net.ListenTCP("tcp6", s.BindAddr)
	if err != nil {
		return fmt.Errorf("Error binding: %q", err.Error())
	}

	go s.HandleConnError()
	for {
		conn, err := s.Listener.AcceptTCP()
		if err != nil {
			log.Printf("Error accepting connection: %q\n", err.Error())
		}

		go s.HandleConnection(conn)
	}

	return nil
}

func main() {
	s := NewServer(":8080")
	s.ConnPool = connection.NewConnectionPool(5, 256)
	if err := s.ServeHTTP(); err != nil {
		panic(err)
	}
}
