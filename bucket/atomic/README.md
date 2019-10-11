如果请求数大于限制,往channel里写数据,协程阻塞

每s从channel 里取下数据,chanel 恢复,同时统计数据清零

