package main

import (
	"errors"
)

import (
	log "github.com/pengswift/gamelibs/nsq-logger"
	. "github.com/pengswift/gamelibs/services/proto"
)

import (
	. "types"
)

var (
	ERROR_STREAM_NOT_OPEN = errors.New("stream not opened yet")
)

// forward message to game server
// 传送消息到游戏服务器
func forward(sess *Session, p []byte) error {
	frame := &Game_Frame{
		Type:    Game_Message,
		Message: p,
	}

	// 检查传输流是否打开
	if sess.Stream == nil {
		return ERROR_STREAM_NOT_OPEN
	}

	//向game服务器传递数据
	if err := sess.Stream.Send(frame); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
