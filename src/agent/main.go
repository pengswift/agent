package main

import (
	"encoding/binary"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

import (
	log "github.com/pengswift/gamelibs/nsq-logger"
)

import (
	. "types"
	"utils"
)

const (
	_port = ":8888"
)

const (
	SERVICE = "[AGENT]"
)

//游戏入口
func main() {
	defer utils.PrintPanicStack()
	go func() {
		log.Info(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	log.SetPrefix(SERVICE)

	tcpAddr, err := net.ResolveTCPAddr("tcp4", _port)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	log.Info("listening on:", listener.Addr())

	// loop accepting
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Warning("accept failed:", err)
			continue
		}
		go handleClient(conn)

		// check server close signal
		select {
		case <-die:
			listener.Close()
			goto FINAL
		default:
		}
	}
FINAL:
	// server closed, wait forever
	// 此处为什么要等1秒
	// 此处等待1秒钟，只是为了wg.Wait() 可以执行到，从而阻塞主线程
	for {
		<-time.After(time.Second)
	}
}

// start a goroutine when a new connection is accepted
func handleClient(conn *net.TCPConn) {
	defer utils.PrintPanicStack()
	// set per-connection socket buffer
	//设置conn连接接受缓存的大小, 当超过缓存时, 会进入阻塞状态,等待被读取
	conn.SetReadBuffer(SO_RCVBUF)

	// set initial socket buffer
	//设置conn连接发送缓存的大小, 当超过缓存时, 会进入阻塞状态,等待被发送成功
	conn.SetWriteBuffer(SO_SNDBUF)

	// initial network control struct
	// 初始化2个字节数组, 用于存储header长度, 既后面要读取的文件长度
	header := make([]byte, 2)
	// 输入流通道, 解析后的数据将放入，等待被处理
	in := make(chan []byte)

	//设置延迟函数，当玩家断开连接时, 函数退出之前，关闭输入流
	defer func() {
		close(in) // session will close
	}()

	// create a new session object for the connection
	//创建session对象, 用于封装客户端和服务器的信息交换
	var sess Session
	host, port, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		log.Error("cannot get remote address:", err)
		return
	}
	//存储用户ip
	sess.IP = net.ParseIP(host)
	//打印用户的ip和端口, 用户可能会双开？
	log.Infof("new connection from:%v port:%v", host, port)

	// session die signal
	sess_die := make(chan bool)

	//SESSION_DIE 监控有问题.................

	// create a write buffer
	// 创建写入buffer对象
	out := new_buffer(conn, sess_die)
	go out.start()

	// start one agent for handling packet
	//记录goroutine个数，让系统接收到关闭命令后，会阻塞主线程，至少所有agent线程退出，已保证数据落地
	wg.Add(1)
	go agent(&sess, in, out, sess_die)

	//network loop
	for {
		// solve dead line problem
		// 设置读超时时间, 如果在任意一次执行Read syscall 返回的时候，超过这个时间点， 则算超时
		conn.SetReadDeadline(time.Now().Add(TCP_READ_DEADLINE * time.Second))
		//先读取2个字节头文件长度
		n, err := io.ReadFull(conn, header)
		if err != nil {
			log.Warningf("read header failed, ip:%v reason:%v size:%v", sess.IP, err, n)
			return
		}
		//将2个字节数组转成int16类型, 不丢失精度
		size := binary.BigEndian.Uint16(header)

		// alloc a byte slice for reading
		// 创建一个指定长度的切片，用于存放具体内容
		payload := make([]byte, size)
		//read msg
		n, err = io.ReadFull(conn, payload)
		if err != nil {
			log.Warningf("read payload failed, ip:%v reason:%v size:%v", sess.IP, err, n)
			return
		}

		select {
		//接收的数据，转入in通道
		case in <- payload: //payload queued
			//监听sess_die 通道
		case <-sess_die:
			log.Warning("connection closed by logic, flag:%v ip:%v", sess.Flag, sess.IP)
			return
		}
	}

	// 好像没有处理连接超时, 如果玩家连上游戏后，一直未操作，再次链接时，会是新的连接？难道客户端在许久没有操作的情况下，先发一次ping, 如果有响应，继续操作，如果没响应，则执行重连？

	// 如果玩家已经退出了游戏，但是通过非正常途径退出的，这时，服务器还保留着该session, 当玩家再次登陆时， 原先的连接何时删除

}

func checkError(err error) {
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
}
