package main

import (
	"time"
)

import (
	spp "github.com/pengswift/gamelibs/services/proto"
)

import (
	. "types"
	"utils"
)

func agent(sess *Session, in chan []byte, out *Buffer, sess_die chan bool) {
	defer wg.Done()
	defer utils.PrintPanicStack()

	// init session
	//存储玩家异步消息, 让消息超过时, 会阻塞，等待等读取,
	//那么如果一直是满的时，玩家好久没上线
	//别的玩家发过来的消息会一直发送不过去, 这时会长期阻塞，怎么解决, 发送方设置超时,当许久没发送成功后，弹出对方信箱已满?
	//  ...... 这里是异步消息，还是离线消息？
	sess.MQ = make(chan []byte, DEFAULT_MQ_SIZE)
	sess.ConnectTime = time.Now()
	sess.LastPacketTime = time.Now()

	// minute timer
	min_timer := time.After(time.Minute)

	// cleanup work
	defer func() {
		close(sess_die)
		//  当session关闭时，连接游戏后段的流需要关闭
		if sess.Stream != nil {
			sess.Stream.CloseSend()
		}
	}()

	// >> the main message loop <<
	for {
		select {
		case msg, ok := <-in: // packet from network
			if !ok {
				return
			}

			//数据包个数++
			sess.PacketCount++
			//更新接收数据包的时间
			sess.PacketTime = time.Now()

			//处理用户请求
			if result := proxy_user_request(sess, msg); result != nil {
				//将结果发送出去
				out.send(sess, result)
			}
			//当处理完之后，将数据包时间设置成上个数据包时间
			sess.LastPacketTime = sess.PacketTime
			//如果异步通道有消息? (内部发送过来的消息）
		case frame := <-sess.MQ:
			switch frame.Type {
			case spp.Game_Message:
				out.send(sess, frame.Message)
				//如果是关闭命令, 则设置关闭状态， (用于后台管理)
			case spp.Game_Kick:
				sess.Flag |= SESS_KICKED_OUT
			}
		case <-min_timer: //minutes timer
			//定时器工作, 用于检测用户行为，防治玩家作弊
			timer_work(sess, out)
			min_timer = time.After(time.Minute)
		case <-die: // server is shuting down...
			sess.Flag |= SESS_KICKED_OUT
		}

		// see if the player should be kicked out.
		if sess.Flag&SESS_KICKED_OUT != 0 {
			return
		}
	}
}
