package sockerserver

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

var handler func(Msg) Msg

/*
 *  one conn will start 3 goroutine : readroutine writeroutine handleroutine
 */
func StartServer(serverAddr string) bool {
	/*tcpAddr, err := net.ResolveIPAddr("ip4", serverAddr)
	if err != nil {
		fmt.Printf("%s %v\n", serverAddr, err)
		return false
	}*/
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		fmt.Printf("%v", err)
		return false
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("%v", err)
			continue
		}
		go handleConn(conn)
	}
	return true
}

type ConnContext struct {
	conn       net.Conn
	quitChan   chan int
	handleChan chan Msg
	writeChan  chan Msg
}

func handleConn(conn net.Conn) {
	fmt.Printf("new conn\n")
	loopers := []func(*ConnContext){readLoop, writeLoop, handleLoop}
	connCtx := new(ConnContext)
	connCtx.conn = conn
	connCtx.quitChan = make(chan int, 1)
	connCtx.handleChan = make(chan Msg, 1000)
	connCtx.writeChan = make(chan Msg, 1000)
	for _, l := range loopers {
		looper := l
		go looper(connCtx)
	}
}

/*
 *msg pattern: 0x6666|data length|data
 */
const (
	MSG_HEAD       = 0x6666
	MSG_HEAD_BYTES = 2
	DATA_LEN_BYTES = 4
)
const (
	READ_MSG_HEAD = 1
	READ_DATA_LEN = 2
	READ_DATA     = 3
)

type Msg struct {
	conn net.Conn
	data []byte
}

func readLoop(connCtx *ConnContext) {
	status := READ_MSG_HEAD
	bufferReader := bufio.NewReader(connCtx.conn)
	var recvBuffer []byte
	tmpBuffer := make([]byte, 1024)
	var dataLen uint32
	needReadMore := false
	for {
		select {
		case <-connCtx.quitChan:
			return
		default:
		}
		n, err := bufferReader.Read(tmpBuffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("client closed\n")
			}
			//todo : send this to other loop by channel
			return
		}
		recvBuffer = append(recvBuffer, tmpBuffer[:n]...)
		needReadMore = false
		for needReadMore == false {
			fmt.Printf("status %v buffer %v buffer len %v\n", status, recvBuffer, len(recvBuffer))
			switch status {
			case READ_MSG_HEAD:
				if len(recvBuffer) >= MSG_HEAD_BYTES {
					if recvBuffer[0] == 0x66 && recvBuffer[1] == 0x66 {
						status = READ_DATA_LEN
					} else {
						fmt.Printf("msg head is broken")
						return
					}
				} else {
					needReadMore = true
				}
			case READ_DATA_LEN:
				if len(recvBuffer) >= MSG_HEAD_BYTES+DATA_LEN_BYTES {
					bufReader := bytes.NewReader(recvBuffer[MSG_HEAD_BYTES : MSG_HEAD_BYTES+DATA_LEN_BYTES])
					binary.Read(bufReader, binary.BigEndian, &dataLen)
					status = READ_DATA
				} else {
					needReadMore = true
				}
			case READ_DATA:
				if len(recvBuffer) >= int(MSG_HEAD_BYTES+DATA_LEN_BYTES+dataLen) {
					headLen := uint32(MSG_HEAD_BYTES + DATA_LEN_BYTES)
					var msg Msg
					msg.conn = connCtx.conn
					msg.data = append(msg.data, recvBuffer[headLen:headLen+dataLen]...)
					connCtx.handleChan <- msg
					var tmp []byte
					tmp = append(tmp, recvBuffer[headLen+dataLen:]...)
					recvBuffer = tmp
					status = READ_MSG_HEAD
					needReadMore = true
				} else {
					needReadMore = true
				}
			}
		}
	}
}
func writeLoop(connCtx *ConnContext) {
	for {
		select {
		case msg := <-connCtx.writeChan:
			sendData := make([]byte, MSG_HEAD_BYTES+DATA_LEN_BYTES+len(msg.data))
			sendData = append(sendData, 0x66, 0x66)
			dataLen := uint32(len(msg.data))
			binary.BigEndian.PutUint32(sendData[MSG_HEAD_BYTES:MSG_HEAD_BYTES+DATA_LEN_BYTES], dataLen)
			sendData = append(sendData, msg.data...)
			connCtx.conn.Write(sendData)
		case <-connCtx.quitChan:
			return
		}
	}

}
func handleLoop(connCtx *ConnContext) {
	for {
		select {
		case msg := <-connCtx.handleChan:
			resultMsg := handler(msg)
			if len(resultMsg.data) > 0 {
				connCtx.writeChan <- resultMsg
			}
		case <-connCtx.quitChan:
			return
		}
	}
}
func RegisterHandler(h func(Msg) Msg) {
	handler = h
}
