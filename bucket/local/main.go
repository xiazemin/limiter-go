package ratelimit

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// todo timer 需要 stop
type localCounterLimiter struct {
	limit Limit

	limitCount int32 // 内部使用，对 limit.count 做了 <0 时的转换

	ticker *time.Ticker
	quit chan bool

	lock sync.Mutex
	newTerm *sync.Cond
	count int32
}

func (lim *localCounterLimiter) init() {
	lim.newTerm = sync.NewCond(&lim.lock)
	lim.limitCount = lim.limit.Count()

	if lim.limitCount < 0 {
		lim.limitCount = math.MaxInt32 // count 永远不会大于 limitCount，后面的写法保证溢出也没问题
	} else if lim.limitCount == 0  {
		// 禁止访问, 会无限阻塞
	} else {
		lim.ticker = time.NewTicker(lim.limit.Period())
		lim.quit = make(chan bool, 1)

		go func() {
			for {
				select {
				case <- lim.ticker.C:
					fmt.Println("ticker .")
					atomic.StoreInt32(&lim.count, 0)
					lim.newTerm.Broadcast()

				//lim.newTerm.L.Unlock()
				case <- lim.quit:
					fmt.Println("work well .")
					lim.ticker.Stop()
					return
				}
			}
		}()
	}
}

// todo 需要机制来防止无限阻塞, 不超时也应该有个极限时间
func (lim *localCounterLimiter) Acquire() error {
	if lim.limitCount == 0 {
		return errors.New("rate limit is 0, infinity wait")
	}

	lim.newTerm.L.Lock()
	for lim.count >= lim.limitCount {
		// block instead of spinning
		lim.newTerm.Wait()
		//fmt.Println(count, lim.limitCount)
	}
	lim.count++
	lim.newTerm.L.Unlock()

	return nil
}

func (lim *localCounterLimiter) TryAcquire() bool {
	count := atomic.AddInt32(&lim.count, 1)
	if count > lim.limitCount {
		return false
	} else {
		return true
	}
}