package redislockv2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

// var (
// 	defaultExpireTime = 5 * time.Second //单位：s
// )

var ErrLockAcquiredByOthers = errors.New("lock is acquired by others")

type RedisLock struct {
	LockOptions
	key      string
	Id       string //锁的标识
	redisCli *redis.Client

	// 停止看门狗
	stopDog context.CancelFunc //通过context.CancelFunc去停止看门狗
}

// func NewRedisLock(cli *redis.Client, key string) *RedisLock {
// 	id := strings.Join(strings.Split(uuid.New().String(), "-"), "")
// 	return &RedisLock{
// 		key:      key,
// 		expire:   defaultExpireTime,
// 		Id:       id,
// 		redisCli: cli,
// 	}
// }

func NewRedisLock(cli *redis.Client, key string, opts ...LockOptionFunc) *RedisLock {
	id := strings.Join(strings.Split(uuid.New().String(), "-"), "")
	lock := RedisLock{
		key:      key,
		Id:       id,
		redisCli: cli,
	}
	//执行一些配置操作
	for _, optFunc := range opts {
		optFunc(&lock.LockOptions)
	}

	setLock(&lock.LockOptions)
	return &lock
}

//	func (lock *RedisLock) Lock() (bool, error) {
//		return lock.redisCli.SetNX(lock.key, lock.Id, lock.expire).Result()
//	}
func (lock *RedisLock) Lock() (success bool, err error) {
	defer func() {
		if success && err == nil {
			lock.watchDog()
		}
	}()

	//不管是否是阻塞的，都是要先获取一次锁
	success, err = lock.tryLock()
	if success && err == nil {
		return success, err
	}

	//非阻塞加锁失败的话，直接返回错误
	if !lock.isBlock {
		return false, err
	}

	//基于阻塞模式轮询去获取锁
	success, err = lock.blockingLock()
	return
}

func (lock *RedisLock) tryLock() (bool, error) {
	return lock.redisCli.SetNX(lock.key, lock.Id, lock.expire).Result()
}

func (lock *RedisLock) blockingLock() (bool, error) {
	timeoutCh := time.After(lock.blockWaitingTime)
	//轮询ticker，定时器, 100ms循环一次去获取锁
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCh:
			return false, fmt.Errorf("block waiting timeout,err:%w", ErrLockAcquiredByOthers)
		case <-ticker.C:
			success, err := lock.tryLock() //尝试获取锁
			if success && err == nil {
				return success, nil
			}
		}
	}
}

func (lock *RedisLock) Unlock() error {
	defer func() {
		//停止看门狗
		if lock.stopDog != nil {
			lock.stopDog()
		}
	}()

	script := redis.NewScript(LauCheckAndDelete)
	result, err := script.Run(lock.redisCli, []string{lock.key}, lock.Id).Int64()
	if err != nil {
		return err
	}
	if result != 1 {
		return errors.New("can not unlock without ownership of lock")
	}
	return nil
}

// =============看门狗的操作
func (lock *RedisLock) watchDog() {
	if !lock.wathchDogMode {
		return
	}

	//启动看门狗
	var ctx context.Context
	ctx, lock.stopDog = context.WithCancel(context.Background())
	go func() {
		lock.runWatchDog(ctx)
	}()
}

func (lock *RedisLock) runWatchDog(ctx context.Context) error {
	//开启一个定时器
	ticker := time.NewTicker(lock.watchDogWorkStepTime)
	defer ticker.Stop()
	script := redis.NewScript(LauCheckThenExpire)
	for {
		select {
		case <-ticker.C:
			result, err := script.Run(lock.redisCli, []string{lock.key}, lock.Id, lock.expire+3).Result()
			if err != nil {
				return err
			}
			if ret, _ := result.(int64); ret != 1 {
				return errors.New("can not expire lock without ownership of lock")
			}
		case <-ctx.Done():
			return nil
		}
	}
}
