package main

import (
	"sync"
	"time"
)

//LimitRate 限速
type LimitRate struct {
	rate     int
	begin    time.Time
	count    int
	lock     sync.Mutex
}

//Limit Limit
func (l *LimitRate) Limit() bool {
	result := true
	l.lock.Lock()
	//达到每秒速率限制数量，检测记数时间是否大于1秒
	//大于则速率在允许范围内，开始重新记数，返回true
	//小于，则返回false，记数不变
	if l.count == l.rate {
		if time.Now().Sub(l.begin) >= time.Second {
			//速度允许范围内，开始重新记数
			l.begin = time.Now()
			l.count = 0
		} else {
			result = false
		}
	} else {
		//没有达到速率限制数量，记数加1
		l.count++
	}
	l.lock.Unlock()

	return result
}

//SetRate 设置每秒允许的请求数
func (l *LimitRate) SetRate(r int) {
	l.rate = r
	l.begin = time.Now()
}

//GetRate 获取每秒允许的请求数
func (l *LimitRate) GetRate() int {
	return l.rate
}
