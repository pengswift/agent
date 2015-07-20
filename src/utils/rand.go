package utils

import (
	"time"
)

var (
	x0  uint32      = uint32(time.Now().UnixNano()) //纳米时间戳
	a   uint32      = 1664525                       //
	c   uint32      = 1013904223                    //
	LCG chan uint32                                 //4个字节长int类型的管道
)

const (
	PRERNG = 1024 // 管道容量
)

//全局快速随机数生成器
func init() {
	LCG = make(chan uint32, PRERNG)
	go func() {
		for {
			x0 = a*x0 + c
			LCG <- x0
		}
	}()
}
