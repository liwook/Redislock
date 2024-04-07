package redislockv2

import (
	"time"
)

const (
	// 默认的分布式锁过期时间
	DefaultExpireTime = 20 * time.Second
	// 看门狗工作时间间隙
	DefaultWatchDogWorkStepTime = 10 * time.Second
	DefaultBlockWaitingTime     = 8 * time.Second
)

type LockOptions struct {
	isBlock              bool
	blockWaitingTime     time.Duration
	expire               time.Duration
	wathchDogMode        bool
	watchDogWorkStepTime time.Duration
}

type LockOptionFunc func(*LockOptions)

// 设置阻塞等待
func WithBlock() LockOptionFunc {
	return func(option *LockOptions) {
		option.isBlock = true
	}
}

// 设置阻塞等待时间的上限
func WithBlockWaiting(waiting time.Duration) LockOptionFunc {
	return func(option *LockOptions) {
		option.blockWaitingTime = waiting
	}
}

// 设置续期的时长,也是key过期的时长
func WithExpire(exprie time.Duration) LockOptionFunc {
	return func(option *LockOptions) {
		option.expire = exprie
	}
}

// 设置看门狗的工作间隔时长，即是每次去查看锁是否过期的工作时间间隔
func WithWatchDogWorkStepTime(watch time.Duration) LockOptionFunc {
	return func(option *LockOptions) {
		option.watchDogWorkStepTime = watch
	}
}

func setLock(o *LockOptions) {
	if o.isBlock && o.blockWaitingTime <= 0 {
		//没有设置默认阻塞时间就使用默认阻塞时长
		o.blockWaitingTime = DefaultBlockWaitingTime
	}
	if o.watchDogWorkStepTime == 0 {
		o.watchDogWorkStepTime = DefaultWatchDogWorkStepTime
	}

	//简单起见，就设置是开启看门狗模式
	o.wathchDogMode = true

	if o.expire == 0 {
		o.expire = DefaultExpireTime
	}
	//比较续期时长和看门狗工作时间间隔
	if o.expire <= o.watchDogWorkStepTime {
		o.watchDogWorkStepTime = o.expire - 2
	}
}

// 红锁的操作
type RedLockOptionFunc func(*RedLockOptions)

type RedLockOptions struct {
	singleNodeTimeout time.Duration //单个节点的请求耗时的超时时间
	exprie            time.Duration //整个红锁的过期时间
}

func WithSingleNodeTimeout(singleNodeTimeout time.Duration) RedLockOptionFunc {
	return func(opt *RedLockOptions) {
		opt.singleNodeTimeout = singleNodeTimeout
	}
}

func WithRedLockExpire(expire time.Duration) RedLockOptionFunc {
	return func(opt *RedLockOptions) {
		opt.exprie = expire
	}
}

type SingleNode struct {
	Address  string
	Password string
}

func setRedLock(opt *RedLockOptions) {
	if opt.singleNodeTimeout <= 0 {
		opt.singleNodeTimeout = DefaultSingleLockTimeout
	}

	if opt.exprie <= 0 {
		opt.exprie = DefaultExpireTime
	}
}
