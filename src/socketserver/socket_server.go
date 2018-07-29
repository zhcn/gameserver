package sockerserver

import (
	"net"
)

var handler func(net.Conn, []byte) []byte

/*
 *  one conn will start 3 goroutine : readroutine writeroutine handleroutine
 */
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
		go handleConn(conn)
	}
	return true
}

type ConnContext struct {
	conn       net.Conn
	quitChan   chan int
	handleChan chan Msg
	writeChan  chan []byte
}

func handleConn(conn net.Conn) {
	loopers := []func(net.Conn){readLoop, writeLoop, handleLoop}
	connCtx := new(ConnContext)
	connCtx.connClose = make(chan int, 1)
	connCtx.handleChan = make(chan Msg, 1000)
	connCtx.writeChan = make(chan []byte)
	for _, l := range loopers {
		looper := l
		go looper(conn, connCtx)
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

func readLoop(connCtx *connContext) {
	status := READ_MSG_HEAD
	bufferReader := bufio.NewReader(connCtx.conn)
	var recvBuffer []byte
	var dataLen uint32
	for {
		select {
		case <-connCtx.quitChan:
			return
		default:
		}
		var tmpBuffer []byte
		n, err := bufferReader.Read(tmpBuffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("client closed\n")
			}
			//todo : send this to other loop by channel
			return
		}
		recvBuffer = append(recvBuffer, tmpBuffer...)
		switch status {
		case READ_MSG_HEAD:
			if len(recvBuffer >= MSG_HEAD_BYTES) {
				if recvBuffer[0] == 0x66 && recvBuffer[1] == 0x66 {
					status = READ_DATA_LEN
				} else {
					fmt.Printf("msg head is broken")
					return
				}
			}
			break
		case READ_DATA_LEN:
			if len(recvBuffer) >= MSG_HEAD_BYTES+DATA_LEN_BYTES {
				binary.Read(recvBuffer[MSG_HEAD_BYTES:MSG_HEAD_BYTES+DATA_LEN_BYTES], binary.BigEndian, &dataLen)
				status = READ_DATA
			}
			break
		case READ_DATA:
			if len(recvBuffer) >= MSG_HEAD_BYTES+DATA_LEN_BYTES+dataLen {
				headLen := MSG_HEAD_BYTES + DATA_LEN_BYTES
				var msg Msg
				msg.conn = conn
				msg.data = recvBuffer[headLen : headLen+dataLen]
				connCtx.handleChan <- msg
				var tmp []byte
				tmp = append(tmp, recvBuffer[headLen+dataLen:]...)
				recvBuffer = tmp
				status = READ_MSG_HEAD
			}
			break
		}
	}
}
func writeLoop(connCtx *connContext) {
	for {
		select {
		case msg := <-connCtx.writeChan:
			sendData := make([]byte, MSG_HEAD_BYTES+DATA_LEN_BYTES+len(msg))
			sendData = append(sendData, 0x66, 0x66)
			dataLen := uint32(len(msg.data))
			binary.Write(sendData[MSG_HEAD_BYTES:MSG_HEAD_BYTES+DATA_LEN_BYTES], binary.BigEndian, dataLen)
			sendData = append(sendData, msg...)
			connCtx.conn.Write(sendData)
		case connCtx.quitChan:
			return
		}
	}

}
func handleData(connCtx *connContext) {
	for {
		select {
		case msg := <-connCtx.handleChan:
			resultMsg := handler(connCtx.conn, msg)
			if len(resultMsg) > 0 {
				connCtx.writeChan <- resultMsg
			}
		case <-connCtx.quitChan:
			return
		}
	}
}
func RegisterHandler(h func(net.Conn, []byte) []byte) {
	handler = h
}
