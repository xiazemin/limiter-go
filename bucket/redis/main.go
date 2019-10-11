package ratelimit

import (
	"math"
	"sync"
	"xg-go/log"
	"xg-go/xg/common"
)

type redisCounterLimiter struct {
	limit      DistLimit
	limitCount int32 // 内部使用，对 limit.count 做了 <0 时的转换

	redisClient *common.RedisClient

	once sync.Once // 退化为本地计数器的时候使用
	localLim Limiter

			 //script string
}

func (lim *redisCounterLimiter) init() {
	lim.limitCount = lim.limit.Count()
	if lim.limitCount < 0 {
		lim.limitCount = math.MaxInt32
	}

	//lim.script = buildScript()
}

//func buildScript() string {
//  sb := strings.Builder{}
//
//  sb.WriteString("local c")
//  sb.WriteString("\nc = redis.call('get',KEYS[1])")
//  // 调用不超过最大值，则直接返回
//  sb.WriteString("\nif c and tonumber(c) > tonumber(ARGV[1]) then")
//  sb.WriteString("\nreturn c;")
//  sb.WriteString("\nend")
//  // 执行计算器自加
//  sb.WriteString("\nc = redis.call('incr',KEYS[1])")
//  sb.WriteString("\nif tonumber(c) == 1 then")
//  sb.WriteString("\nredis.call('expire',KEYS[1],ARGV[2])")
//  sb.WriteString("\nend")
//  sb.WriteString("\nif tonumber(c) == 1 then")
//  sb.WriteString("\nreturn c;")
//
//  return sb.String()
//}

func (lim *redisCounterLimiter) Acquire() error {
	panic("implement me")
}

func (lim *redisCounterLimiter) TryAcquire() (success bool) {
	defer func() {
		// 一般是 redis 连接断了，会触发空指针
		if err := recover(); err != nil {
			//log.Errorw("TryAcquire err", common.ERR, err)
			//success = lim.degradeTryAcquire()
			//return
			success = true
		}

		// 没有错误，判断是否开启了 local 如果开启了，把它停掉
		//if lim.localLim != nil {
		//  // stop 线程安全
		//  lim.localLim.Stop()
		//}
	}()

	count, err := lim.redisClient.IncrBy(lim.limit.Key(), 1)
	//panic("模拟 redis 出错")
	if err != nil {
		log.Errorw("TryAcquire err", common.ERR, err)
		panic(err)
	}

	// *2 是为了保留久一点，便于观察
	err = lim.redisClient.Expire(lim.limit.Key(), int(2 * lim.limit.Period().Seconds()))
	if err != nil {
		log.Errorw("TryAcquire error", common.ERR, err)
		panic(err)
	}

	// 业务正确的情况下 确认超限
	if int32(count) > lim.limitCount {
		return false
	}

	return true

	//keys := []string{lim.limit.Key()}
	//
	//log.Errorw("TryAcquire ", keys, lim.limit.Count(), lim.limit.Period().Seconds())
	//count, err := lim.redisClient.Eval(lim.script, keys, lim.limit.Count(), lim.limit.Period().Seconds())
	//if err != nil {
	//  log.Errorw("TryAcquire error", common.ERR, err)
	//  return false
	//}
	//
	//
	//typeName := reflect.TypeOf(count).Name()
	//log.Errorw(typeName)
	//
	//if count != nil && count.(int32) <= lim.limitCount {
	//
	//  return true
	//}
	//return false
}

func (lim *redisCounterLimiter) Stop() {
	// 判断是否开启了 local 如果开启了，把它停掉
	if lim.localLim != nil {
		// stop 线程安全
		lim.localLim.Stop()
	}
}

func (lim *redisCounterLimiter) degradeTryAcquire() bool {
	lim.once.Do(func() {
		count := lim.limit.Count() / lim.limit.ClusterNum()
		limit := LocalLimit {
			name: lim.limit.Name(),
			key: lim.limit.Key(),
			count: count,
			period: lim.limit.Period(),
			limitType: lim.limit.LimitType(),
		}

		lim.localLim = NewLimiter(&limit)
	})

	return lim.localLim.TryAcquire()
}