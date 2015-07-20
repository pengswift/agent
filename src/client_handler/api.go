package client_handler

import (
	"misc/packet"
	. "types"
)

var Code = map[string]int16{
	"heart_beat_req":         0,  //心跳包..
	"heart_beat_ack":         1,  //心跳包回复
	"user_login_req":         10, //登录
	"user_login_success_ack": 11, //登陆成功
	"user_login_faild_ack":   12, //登陆失败
	"client_error_ack":       13, //客户端错误
	"get_seed_req":           30, //socket通讯加密使用
	"get_seed_ack":           31, //socket通讯加密使用
}

var RCode = map[int16]string{
	0:  "heart_beat_req",         //心跳包..
	1:  "heart_beat_ack",         //心跳包回复
	10: "user_login_req",         //客户端发送登陆请求
	11: "user_login_success_ack", //登陆成功
	12: "user_login_faild_ack",   //登陆失败
	13: "client_error_ack",       //客户端错误
	30: "get_seed_req",           //socket通讯加密使用
	31: "get_seed_ack",           //socket通讯加密使用
}

var Handlers map[int16]func(*Session, *packet.Packet) []byte

func init() {
	Handlers = map[int16]func(*Session, *packet.Packet) []byte{
		0:  P_heart_beat_req,
		10: P_user_login_req,
		30: P_get_seed_req,
	}
}
