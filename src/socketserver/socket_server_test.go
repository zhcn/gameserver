package sockerserver

import (
	"encoding/binary"
	"fmt"
	"net"
	"testing"
)

func testHandler(msg Msg) Msg {
	fmt.Printf("handler %v\n", msg.data)
	return msg
}

func client() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:12345")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	buf := make([]byte, 1024)
	buf[0] = 0x66
	buf[1] = 0x66
	binary.BigEndian.PutUint32(buf[2:], 1)
	buf[6] = 0x06
	fmt.Printf("send buf %v\n", buf[:7])
	conn.Write(buf[:7])
	var ch chan int
	<-ch
}

func TestStartServer(t *testing.T) {
	go StartServer("127.0.0.1:12345")
	RegisterHandler(testHandler)
	client()
	var test chan int
	<-test
}
