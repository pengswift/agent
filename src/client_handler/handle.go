package client_handler

import (
	//"crypto/rc4"
	//"fmt"
	log "github.com/pengswift/libs/nsq-logger"
	//"math/big"
)

import (
	//"misc/crypto/dh"
	"misc/packet"
	. "types"
)

// 心跳包
func P_heart_beat_req(sess *Session, reader *packet.Packet) []byte {
	return packet.Pack(Code["heart_beat_ack"], nil, nil)
}

// 密钥交换
func P_get_seed_req(sess *Session, reader *packet.Packet) []byte {
	//tbl, _ := PKT_seed_info(reader)

	//// KEY1
	//X1, E1 := dh.DHExchange()
	//KEY1 := dh.DHKey(X1, big.NewInt(int64(tbl.F_client_send_seed)))

	//// KEY2
	//X2, E2 := dh.DHExchange()
	//KEY2 := dh.DHKey(x2, big.NewInt(int64(tbl.F_client_receive_seed)))

	//ret := seed_info{int32(E1.Int64()), int32(E2.Int64())}
	////服务器加密种子是客户端解密种子
	//encoder, err := rc4.NewCliper([]byte(fmt.Sprintf("%v%v", SALT, KEY2)))
	//if err != nil {
	//	log.Critical(err)
	//	return nil
	//}
	//decoder, err := rc4.NewCliper([]byte(fmt.Sprintf("%v%v", SALT, KEY1)))
	//if err != nil {
	//	log.Critical(err)
	//	return nil
	//}

	//sess.Encoder = encoder
	//sess.Decoder = decoder
	//sess.Flag |= SESS_KEYEXCG
	//return packet.Pack(Code["get_seed_ack"], ret, nil)
	return nil
}

// 玩家登陆过程
func P_user_login_req(sess *Session, reader *packet.Packet) []byte {
	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Error(err)
		panic("error occured in protocol module")
	}
}
