package main

import (
	"fmt"
	"os"
	"time"
)

import (
	"github.com/peterbourgon/g2s"
)

import (
	log "github.com/pengswift/gamelibs/nsq-logger"
)

import (
	"client_handler"
	"misc/packet"
	. "types"
	"utils"
)

const (
	STATSD_PREFIX       = "API.NR"
	ENV_STATSD          = "STATSD_HOST"
	DEFAULT_STATSD_HOST = "127.0.0.1:8125"
)

var (
	_statter g2s.Statter
)

func init() {
	addr := DEFAULT_STATSD_HOST
	if env := os.Getenv(ENV_STATSD); env != "" {
		addr = env
	}

	s, err := g2s.Dial("udp", addr)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	_statter = s

}

// client protocol handle proxy
func proxy_user_request(sess *Session, p []byte) []byte {
	start := time.Now()
	defer utils.PrintPanicStack()
	//解密
	if sess.Flag&SESS_ENCRYPT != 0 {
		//使用2组密钥匙, 增加破解难度
		sess.Decoder.XORKeyStream(p, p)
	}

	//封装成reader
	reader := packet.Reader(p)

	// 读客户端数据包序列号(1,2,3...)
	// 可避免重放攻击-REPLAY-ATTACK
	//数据包的前4个字节存放的是客户端的发包数量
	//客户端每次发包要包含第几次发包信息
	seq_id, err := reader.ReadU32()
	if err != nil {
		log.Error("read client timestamp failed:", err)
		sess.Flag |= SESS_KICKED_OUT
		return nil
	}

	// 数据包个数验证
	if seq_id != sess.PacketCount {
		//数据包真实长度是，总长度-4个字节的包个数长度-2个字节的协议号长度
		log.Errorf("illegal packet sequeue id:%v should be:%v size:%v", seq_id, sess.PacketCount, len(p)-6)
		sess.Flag |= SESS_KICKED_OUT
		return nil
	}

	// 读协议号
	b, err := reader.ReadS16()
	if err != nil {
		log.Error("read protocol number failed.")
		sess.Flag |= SESS_KICKED_OUT
		return nil
	}

	//根据协议号段 做服务划分
	var ret []byte
	if b > MAX_PROTO_NUM {
		//去除4字节发包个数,将其余的向前传递
		if err := forward(sess, p[4:]); err != nil {
			log.Errorf("service id:%v execute failed, error:%v", b, err)
			sess.Flag |= SESS_KICKED_OUT
			return nil
		}
	} else {
		if h := client_handler.Handlers[b]; h != nil {
			ret = h(sess, reader)
		} else {
			log.Errorf("service id:%v not bind", b)
			sess.Flag |= SESS_KICKED_OUT
			return nil
		}
	}

	// 统计处理时间
	elasped := time.Now().Sub(start)
	if b != 0 { //排除心跳包日志
		log.Trace("[REQ]", client_handler.RCode[b])
		_statter.Timing(1.0, fmt.Sprintf("%v%v", STATSD_PREFIX, client_handler.RCode[b]), elasped)
	}
	return ret
}
