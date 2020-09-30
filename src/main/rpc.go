package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var zRecvCount = uint32(0)
var lRecvCount = uint32(0)
var total = uint32(100000)

var z0 = "吃了没，您吶?"
var z3 = "嗨！吃饱了溜溜弯儿。"
var z5 = "回头去给老太太请安！"
var l1 = "刚吃。"
var l2 = "您这，嘛去？"
var l4 = "有空家里坐坐啊。"

var liWriteLock sync.Mutex
var zhangWriteLock sync.Mutex

type RequestResponse struct {
	Serial  uint32
	Payload string
}

// 序列化RequestResponse，并发送
// 序列化后的结构如下：
//   长度  4字节
//   Serial 4字节
//   PayLoad 变长
func writeTo(r *RequestResponse, conn *net.TCPConn, lock *sync.Mutex) {
	lock.Lock()
	defer lock.Unlock()
	payloadBytes := []byte(r.Payload)

	serialBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(serialBytes, r.Serial)

	length := uint32(len(payloadBytes) + len(serialBytes))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)

	conn.Write(lengthBytes)
	conn.Write(serialBytes)
	conn.Write(payloadBytes)
}

// 接收数据，反序列化成RequestResponse
func readFrom(conn *net.TCPConn) (*RequestResponse, error) {
	ret := &RequestResponse{}
	buf := make([]byte, 4)

	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, fmt.Errorf("读长度故障：%s", err.Error())
	}

	length := binary.BigEndian.Uint32(buf)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, fmt.Errorf("读Serial故障：%s", err.Error())
	}

	ret.Serial = binary.BigEndian.Uint32(buf)

	payloadBytes := make([]byte, length-4)
	if _, err := io.ReadFull(conn, payloadBytes); err != nil {
		return nil, fmt.Errorf("读Payload故障：%s", err.Error())
	}

	ret.Payload = string(payloadBytes)
	return ret, nil
}

// 张大爷的耳朵
func zhangDaYeListen(conn *net.TCPConn, wg *sync.WaitGroup) {
	defer wg.Done()
	for zRecvCount < total*3 {
		r, err := readFrom(conn)
		if err != nil {
			fmt.Print(err.Error())
			break
		}

		if r.Payload == l2 {
			go writeTo(&RequestResponse{r.Serial, z3}, conn, &zhangWriteLock)
		} else if r.Payload == l4 {
			go writeTo(&RequestResponse{r.Serial, z5}, conn, &zhangWriteLock)
		} else if r.Payload == l1 {
			// 不用回复
		} else {
			fmt.Print("张大爷听不懂：" + r.Payload)
			break
		}
		zRecvCount++
	}
}

// 张大爷的嘴
func zhangDaYeSay(conn *net.TCPConn) {
	nextSerial := uint32(0)
	for i := uint32(0); i < total; i++ {
		writeTo(&RequestResponse{nextSerial, z0}, conn, &zhangWriteLock)
		nextSerial++
	}
}

// 李大爷的耳朵，实现是和张大爷类似的
func liDaYeListen(conn *net.TCPConn, wg *sync.WaitGroup) {
	defer wg.Done()
	for lRecvCount < total*3 {
		r, err := readFrom(conn)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		// fmt.Println("李大爷收到：" + r.Payload)
		if r.Payload == z0 { // 如果收到：吃了没，您吶?
			writeTo(&RequestResponse{r.Serial, l1}, conn, &liWriteLock) // 回复：刚吃。
		} else if r.Payload == z3 {
			// do nothing
		} else if r.Payload == z5 {
			// do nothing
		} else {
			fmt.Println("李大爷听不懂：" + r.Payload)
			break
		}
		lRecvCount++
	}
}

// 李大爷的嘴
func liDaYeSay(conn *net.TCPConn) {
	nextSerial := uint32(0)
	for i := uint32(0); i < total; i++ {
		writeTo(&RequestResponse{nextSerial, l2}, conn, &liWriteLock)
		nextSerial++
		writeTo(&RequestResponse{nextSerial, l4}, conn, &liWriteLock)
		nextSerial++
	}
}

func startServer(wg *sync.WaitGroup) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:9999")
	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)
	defer tcpListener.Close()
	fmt.Println("张大爷在胡同口等着 ...")
	for {
		conn, err := tcpListener.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println("碰见一个李大爷:" + conn.RemoteAddr().String())
		go zhangDaYeListen(conn, wg)
		go zhangDaYeSay(conn)
	}
}

func startClient(wg *sync.WaitGroup) *net.TCPConn {
	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", "127.0.0.1:9999")
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	go liDaYeListen(conn, wg)
	go liDaYeSay(conn)
	return conn
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go startServer(&wg)
	time.Sleep(time.Second)
	conn := startClient(&wg)
	t1 := time.Now()
	wg.Wait()
	elapsed := time.Since(t1)
	conn.Close()
	fmt.Println("耗时: ", elapsed)
}
