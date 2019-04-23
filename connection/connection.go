package connection

import (
	"fmt"
	"io"
	"net"
	"sort"
	"strings"

	"github.com/tywkeene/http-rfc/request"
)

type Connection struct {
	Free        int
	Conn        *net.TCPConn
	WriteBuffer []byte
	ReadBuffer  []byte
}

type ConnectionPool struct {
	Connections []*Connection
}

const (
	HeaderMaxSize = 4096
)

func (c *Connection) Close() error {
	var err error
	err = c.Conn.Close()

	c.Conn = nil
	c.Free = 1
	c.WriteBuffer = nil
	c.ReadBuffer = nil

	return err
}

func (c *Connection) WriteResponse(buffer []byte) (int, error) {
	return 0, nil
}

func ReadMethod(input []byte) []byte {
	var output []byte
	for i := 0; i < (len(input) - 1); i++ {
		if input[i] == '\r' && input[i+1] == '\n' {
			break
		} else {
			output = append(output, input[i])
		}
	}
	return output
}

func ReadPath() {}

func ParseHeaders(input [][]byte) map[string]string {
	var headers = make(map[string]string, 0)
	for _, str := range input {
		parsed := strings.Split(string(str), " ")
		fmt.Println(parsed)
		headers[string(parsed[0])] = string(parsed[1])
	}
	return headers
}

func parseLines(input []byte) [][]byte {
	var line []byte
	var arr [][]byte = make([][]byte, 0)
	var offset int
	for i := 0; i < (len(input) - 1); i++ {
		if input[i] == '\r' && input[i+1] == '\n' {
			line = input[offset:i]
			i += 2
			offset = i
			arr = append(arr, line)
			continue
		}
	}
	return arr
}

func (c *Connection) readLine() ([]byte, error) {
	var buffer []byte = make([]byte, HeaderMaxSize)
	n, err := c.Conn.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if err == io.EOF {
		return buffer, nil
	}
	if n >= HeaderMaxSize {
		return nil, fmt.Errorf("Error: Header size exceeds maximum")
	}
	return buffer, nil
}

func (c *Connection) ReadRequest() (*request.Request, error) {

	var err error
	var headerLines []byte

	headerLines, err = c.readLine()
	if err != nil && err != io.EOF {
		return nil, err
	}

	lines := parseLines(headerLines)
	for i, line := range lines {
		fmt.Println(i, string(line))
	}

	/*
		method := ReadMethod(headerLines)
		methodLen := len(method)
		headerLines = headerLines[methodLen:]
		fmt.Println(string(headerLines))
	*/

	headers := ParseHeaders(lines)

	fmt.Println(headers)

	return nil, nil
}

func (p *ConnectionPool) FirstFree() uint64 {
	sort.SliceStable(p.Connections, func(i, j int) bool {
		return p.Connections[i].Free < p.Connections[j].Free
	})
	index := sort.Search(len(p.Connections), func(i int) bool {
		return p.Connections[i].Free >= 0
	})
	return uint64(index)
}

func (p *ConnectionPool) AddConnection(rawConn *net.TCPConn) *Connection {
	index := p.FirstFree()
	p.Connections[index].Free = 0
	p.Connections[index].Conn = rawConn
	return p.Connections[index]
}

func NewConnectionPool(size int, readBufferSize uint32) *ConnectionPool {
	p := &ConnectionPool{
		Connections: make([]*Connection, size),
	}

	for i := 0; i < size; i++ {
		p.Connections[i] = &Connection{
			Free:       1,
			Conn:       nil,
			ReadBuffer: make([]byte, 1),
		}
	}
	return p
}
