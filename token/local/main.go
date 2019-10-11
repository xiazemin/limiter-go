package ratelimit

import (
	"context"
	"golang.org/x/time/rate"
	"math"
)

type localTokenBucketLimiter struct {
	limit Limit

	limiter *rate.Limiter // 直接复用令牌桶的
}

func (lim *localTokenBucketLimiter) init() {
	burstCount := lim.limit.Count()
	if burstLimit, ok := lim.limit.(BurstLimit); ok {
		burstCount = burstLimit.BurstCount()
	}

	count := lim.limit.Count()
	if count < 0 {
		count = math.MaxInt32
	}

	f := float64(count) / lim.limit.Period().Seconds()
	if f < 0 {
		f = float64(rate.Inf) // 无限
	} else if f == 0 {
		panic("为 0 的时候，底层实现有问题")
	}

	lim.limiter = rate.NewLimiter(rate.Limit(f), int(burstCount))
}

func (lim *localTokenBucketLimiter) Acquire() error {
	err := lim.limiter.Wait(context.TODO())
	return err
}

func (lim *localTokenBucketLimiter) TryAcquire() bool {
	return lim.limiter.Allow()
}