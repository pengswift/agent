package main

import (
	"fmt"
	//"os"
	//"time"

	log "github.com/pengswift/libs/nsq-logger"
)

import (
	"client_handler"
	"misc/packet"
	. "types"
	"utils"
)

// client protocol handle proxy
func proxy_user_request(sess *Session, p []byte) []byte {
	//start := time.Now()
	defer utils.PrintPanicStack()
	//解密
	if sess.Flag&SESS_ENCRYPT != 0 {
		sess.Decoder.XORKeyStream(p, p)
	}

	//封装成reader
	reader := packet.Reader(p)

	// 读客户端数据包序列号(1,2,3...)
	// 可避免重放攻击-REPLAY-ATTACK
	seq_id, err := reader.ReadU32()
	if err != nil {
		log.Error("read client timestamp failed:", err)
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

	// 数据包序列号验证
	if seq_id != sess.PacketCount {
		log.Errorf("illegal packet sequeue id:%v should be:%v proto:%v size:%v", seq_id, sess.PacketCount, b, len(p)-6)
		sess.Flag |= SESS_KICKED_OUT
		return nil
	}

	var ret []byte
	if b > MAX_PROTO_NUM { // game协议
		// 透传
		err = forward(sess, p)
		if err != nil {
			log.Error("service id:%v execute failed", b)
			sess.Flag |= SESS_KICKED_OUT
			return nil
		}
	} else { // agent保留协议段 [0, MAX_PROTO_NUM]
		// handle有效性检查
		h := client_handler.Handlers[b]
		if h == nil {
			log.Errorf("service id:%v not bind", b)
			sess.Flag |= SESS_KICKED_OUT
			return nil
		}
		//执行
		fmt.Printf("OK fm.Print\n")
		ret = h(sess, reader)
	}

	// 统计处理事件
	//elasped := time.Now().Sub(start)
	//if b != 0 { //排除心跳包日志
	//	log.Trace("[REQ]", b)
	//	//_statter.Timing(1.0, fmt.Printf("%v%v", STATSD_PREFIX, b), elasped)
	//}
	return ret
}
