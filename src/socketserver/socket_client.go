package sockerserver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

func sendData(conn *net.TCPConn, data []byte) {
	buffer := []byte{0x66, 0x66, 0x00, 0x00, 0x00, 0x00}
	binary.BigEndian.PutUint32(buffer[2:6], uint32(len(data)))
	buffer = append(buffer, data...)
	fmt.Printf("client sendData : %v\n", buffer)
	conn.Write(buffer)
}

func clientReadLoop(conn *net.TCPConn) {
	readBuffer := make([]byte, 1024)
	for {
		n, err := conn.Read(readBuffer)
		if err != nil {
			fmt.Printf("client read error\n")
		}
		fmt.Printf("client read data:%v len: %v\n", readBuffer[:n], n)
	}
}

func clientWriteLoop(conn *net.TCPConn, writeChan chan []byte) {
	for {
		select {
		case msg := <-writeChan:
			sendData(conn, msg)
		}
	}
}

func initClient() {
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
	writeChan := make(chan []byte, 1000)
	go clientWriteLoop(conn, writeChan)
	go clientReadLoop(conn)
	var i int32
	for i = 0; i < 10; i++ {
		data := bytes.NewBuffer([]byte{})
		binary.Write(data, binary.BigEndian, i)
		fmt.Printf("client send data : %v\n", data.Bytes())
		writeChan <- data.Bytes()
	}
}
