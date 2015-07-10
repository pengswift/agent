package main

import (
	log "github.com/pengswift/libs/nsq-logger"
	"time"
)

import (
	. "types"
)

// 玩家一分钟定时器
func timer_work(sess *Session, out *Buffer) {
	// 发包频率控制, 太高的RPS直接踢掉
	interval := time.Now().Sub(sess.ConnectTime).Minutes()
	if interval >= 1 { //登陆时长超过1分钟才开始统计rpm。防脉冲
		rpm := float64(sess.PacketCount) / interval

		if rpm > RPM_LIMIT {
			sess.Flag |= SESS_KICKED_OUT
			log.Error("玩家RPM太高 RPM:", rpm)
			return
		}
	}
}
