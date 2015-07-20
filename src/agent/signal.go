package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

import (
	log "github.com/pengswift/gamelibs/nsq-logger"
)

import (
	"utils"
)

var (
	wg sync.WaitGroup
	// server close signal
	die = make(chan bool)
)

// handle unix signals
func sig_handler() {
	defer utils.PrintPanicStack()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)

	for {
		msg := <-ch
		//当接收到关闭信号时，要等待所有agent协程退出
		switch msg {
		case syscall.SIGTERM: //关闭agent
			close(die)
			log.Info("sigterm received")
			log.Info("waiting for agents close, please wait...")
			wg.Wait()
			log.Info("agent shutdown.")
			os.Exit(0)
		}
	}
}
