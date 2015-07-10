package main

import (
	"time"
)

import (
	. "types"
	"utils"
)

func agent(sess *Session, in chan []byte, out *Buffer, sess_die chan bool) {
	defer wg.Done()
	defer utils.PrintPanicStack()

	// init session
	sess.MQ = make(chan []byte, DEFAULT_MQ_SIZE)
	sess.ConnectTime = time.Now()
	sess.LastPacketTime = time.Now()

	// minute timer
	min_timer := time.After(time.Minute)

	// cleanup work
	defer func() {
		close(sess_die)
	}()

	// >> the main message loop <<
	for {
		select {
		case msg, ok := <-in: // packet from network
			if !ok {
				return
			}

			sess.PacketCount++
			sess.PacketTime = time.Now()

			if result := proxy_user_request(sess, msg); result != nil {
				out.send(sess, result)
			}
			sess.LastPacketTime = sess.PacketTime
		case msg := <-sess.MQ:
			out.send(sess, msg)
		case <-min_timer: //minutes timer
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
