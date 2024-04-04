package redislock

import (
	"errors"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

var (
	defaultExpireTime = 5 * time.Second
)

type RedisLock struct {
	key      string
	expire   time.Duration
	Id       string //锁的标识
	redisCli *redis.Client
}

func NewRedisLock(cli *redis.Client, key string) *RedisLock {
	id := strings.Join(strings.Split(uuid.New().String(), "-"), "")
	return &RedisLock{
		key:      key,
		expire:   defaultExpireTime,
		Id:       id,
		redisCli: cli,
	}
}

func (lock *RedisLock) Lock() (bool, error) {
	return lock.redisCli.SetNX(lock.key, lock.Id, lock.expire).Result()
}

// 解锁， 查看锁是否属于自己和del是原子操作
func (lock *RedisLock) Unlock() error {
	script := redis.NewScript(LauCheckAndDelete)
	res, err := script.Run(lock.redisCli, []string{lock.key}, lock.Id).Int64()
	if err != nil {
		return err
	}
	if res != 1 {
		return errors.New("can not unlock because del result not is one")
	}
	return nil
}

// 锁的误删除实现,非原子操作
// func (lock *RedisLock) Unlock() error {
// 	//获取锁并进行判断该锁是否是自己的
// 	val, err := lock.redisCli.Get(lock.key).Result()
// 	if err != nil {
// 		fmt.Println("lock not exit")
// 		return err
// 	}
// 	if val == "" || val != lock.Id {
// 		return errors.New("lock not belong to myself")
// 	}

// 	//进行删除锁
// 	res, err := lock.redisCli.Del(lock.key).Result()
// 	if err != nil {
// 		return err
// 	}
// 	if res != 1 {
// 		return errors.New("can not unlock because del result not is one")
// 	}
// 	return nil
// }

//基于setnx的锁初步实现， 解锁方法
// func (lock *RedisLock) Unlock() error {
// 	res, err := lock.redisCli.Del(lock.key).Result()
// 	if err != nil {
// 		return err
// 	}
// 	if res != 1 {
// 		return errors.New("can not unlock because del result not is one")
// 	}
// 	return nil
// }
