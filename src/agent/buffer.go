package main

import (
	"encoding/binary"
	"net"
	"time"
)

import (
	log "github.com/pengswift/gamelibs/nsq-logger"
)

import (
	"misc/packet"
	. "types"
	"utils"
)

type Buffer struct {
	ctrl    chan bool    //接收退出信号
	pending chan []byte  //等待处理的数据包
	conn    *net.TCPConn //conn
	cache   []byte       //for combined syscall write 合并系统写入byte对象，节省重复创建开销
}

var (
	// for padding packet, random content
	_padding [PADDING_SIZE]byte
)

func init() {
	go func() { // padding content update procedure
		for {
			for k := range _padding {
				//利用快速生成器填充内容
				//生成32位，填充只有后8位，丢失3字节精度
				_padding[k] = byte(<-utils.LCG)
			}
			log.Info("Padding Updated:", _padding)
			// 300秒更新一次填充值
			<-time.After(PADDING_UPDATE_PERIOD * time.Second)
		}
	}()
}

// packet sending procedure
func (buf *Buffer) send(sess *Session, data []byte) {
	// in case of empty packet
	if data == nil {
		return
	}

	//padding
	//if the size of the data to return is tiny, pad with some random numbers
	//如果当前发送数据长度<规定的长度, 会增加填充8字节随机内容
	//这样客户端 如果收到填充的数据， 该如何处理, 主动减少8个字节， 但依据什么来判断呢?
	if len(data) < PADDING_LIMIT {
		data = append(data, _padding[:]...)
	}

	// encryption
	//如果加密开关已开启,则加密发送的数据
	if sess.Flag&SESS_ENCRYPT != 0 { // encryption is enabled
		sess.Encoder.XORKeyStream(data, data)
		// 如果已经交换了密钥, 则下次开启加密
	} else if sess.Flag&SESS_KEYEXCG != 0 { // key is exchanged, encryption is not yet enabled
		sess.Flag &^= SESS_KEYEXCG
		sess.Flag |= SESS_ENCRYPT
	}

	//queue the data for sending
	// 将数据加入pending管道，等待系统调用
	buf.pending <- data
	return
}

// packet sending goroutine
// 数据包发送协程
func (buf *Buffer) start() {
	defer utils.PrintPanicStack()
	for {
		select {
		//监听数据包管道
		case data := <-buf.pending:
			buf.raw_send(data)
		//监听关闭管道
		//?Important
		//ctrl 指向的是 sess_die, 在main.go里面,已经有select监听了? 此处会不会造成问题
		case <-buf.ctrl: //receive session end signal
			//关闭接收管道
			close(buf.pending)
			//关闭conn连接
			buf.conn.Close()
			return
		}
	}
}

// raw packet encapsulation and put it online
// 向客户端发送数据
func (buf *Buffer) raw_send(data []byte) bool {
	//combine output
	sz := len(data)
	//利用cache,减少创建开销
	//写入2个字节的内容长度
	binary.BigEndian.PutUint16(buf.cache, uint16(sz))
	//copy内容到cache内
	copy(buf.cache[2:], data)

	//write data
	//实际要写入内容为 data长度＋2个字节头长度
	n, err := buf.conn.Write(buf.cache[:sz+2])
	if err != nil {
		log.Warningf("Error send reply data, bytes: %v reason: %v", n, err)
		return false
	}
	return true
}

//暂未使用过
func (buf *Buffer) set_write_buffer(bytes int) {
	buf.conn.SetWriteBuffer(bytes)
}

// create a associated write buffer for a session
func new_buffer(conn *net.TCPConn, ctrl chan bool) *Buffer {
	buf := Buffer{conn: conn}
	buf.pending = make(chan []byte)
	buf.ctrl = ctrl
	buf.cache = make([]byte, packet.PACKET_LIMIT+2) //cache长度为数据包长度＋头长度
	return &buf
}
