package sockerserver

import (
	"net"
)

func StartServer(serverAddr string) bool {
	tcpAddr, err := net.ResolveTCPAddr("ip4", serverAddr)
	if err != nil {
		return false
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return false
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go HandleConn(conn)
	}
	return true
}

func HandleConn(conn net.Conn) {

}
