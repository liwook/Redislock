package redislockv2

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
)

const DefaultSingleLockTimeout = 50 * time.Millisecond

type RedLock struct {
	locks []*RedisLock
	RedLockOptions
}

func NewRedLock(key string, nodes []*SingleNode, opts ...RedLockOptionFunc) (*RedLock, error) {
	//节点个数<3，没有意义
	if len(nodes) < 3 {
		return nil, errors.New("the number of node is less than 3")
	}

	lock := RedLock{}
	for _, opt := range opts {
		opt(&lock.RedLockOptions)
	}

	setRedLock(&lock.RedLockOptions)
	if lock.exprie > 0 && time.Duration(len(nodes))*lock.singleNodeTimeout*10 > lock.exprie {
		// 要求所有节点累计的超时阈值要小于分布式锁过期时间的十分之一
		return nil, errors.New("expire thresholds of single node is too long")
	}

	lock.locks = make([]*RedisLock, 0, len(nodes))
	for _, node := range nodes {
		client := redis.NewClient(&redis.Options{
			Addr:     node.Address,
			Password: node.Password,
		})
		lock.locks = append(lock.locks, NewRedisLock(client, key, WithExpire(lock.exprie)))
	}

	return &lock, nil
}

func (r *RedLock) Lock() (bool, error) {
	//成功加锁的个数
	successNum := 0
	//对每个node尝试加锁
	for _, lock := range r.locks {
		startTime := time.Now()
		success, err := lock.Lock()
		cost := time.Since(startTime)
		if err == nil && success && cost <= r.singleNodeTimeout {
			successNum++
		}
	}

	if successNum < (len(r.locks)>>1)+1 {
		return false, errors.New("lock failed,lock nodes are Not enough for half")
	}
	return true, nil
}

// 解锁，需对所有节点解锁
func (r *RedLock) Unlock() error {
	var allErr error
	for _, lock := range r.locks {
		if err := lock.Unlock(); err != nil {
			if allErr == nil {
				allErr = err
			}
		}
	}
	return allErr
}
