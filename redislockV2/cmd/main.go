package main

import (
	"fmt"
	redislock "redislockV2"

	"sync"
	"time"

	"github.com/go-redis/redis"
)

func NewClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
	})
}

func main() {

	testBlockingLock()
}

func testRedis() {
	client := NewClient()
	defer client.Close()

	err := client.Set("name", "newjack", 0)
	if err.Err() != nil {
		fmt.Println(err.Err())
	}
	fmt.Println(err.String())

	val, e := client.Get("name").Result()
	if e != nil {
		fmt.Println(err)
	}
	fmt.Println("name= ", val)
}

func testBlockingLock() {
	client := NewClient()
	defer client.Close()

	val, _ := client.Ping().Result()
	fmt.Println(val)

	key := "blockLock"
	lock := redislock.NewRedisLock(client, key, redislock.WithBlock(), redislock.WithBlockWaiting(15*time.Second))

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		//尝试获取锁
		if success, err := lock.Lock(); success && err == nil {
			fmt.Println("go BLOCKlock get..")
			time.Sleep(4 * time.Second)
			lock.Unlock()
		}

		wg.Done()
	}()

	//尝试获取锁
	if success, err := lock.Lock(); success && err == nil {
		fmt.Println("main BLOCKlock get...")
		time.Sleep(7 * time.Second)
		lock.Unlock()
	}
	wg.Wait()
}

func testWatchDog() {
	client := NewClient()
	defer client.Close()

	val, _ := client.Ping().Result()
	fmt.Println(val)

	key := "blockLock"
	lock1 := redislock.NewRedisLock(client, key, redislock.WithBlock(), redislock.WithExpire(5*time.Second), redislock.WithBlockWaiting(15*time.Second))
	lock2 := redislock.NewRedisLock(client, key, redislock.WithBlock(), redislock.WithExpire(5*time.Second), redislock.WithBlockWaiting(15*time.Second))

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		//尝试获取锁
		if success, err := lock1.Lock(); success && err == nil {
			fmt.Println("go BLOCKlock get..")
			time.Sleep(4 * time.Second)
			lock1.Unlock()
		}

		wg.Done()
	}()

	//尝试获取锁
	if success, err := lock2.Lock(); success && err == nil {
		fmt.Println("main BLOCKlock get...")
		time.Sleep(7 * time.Second)
		lock2.Unlock()
	}
	wg.Wait()
}

func testReadLock() {
	nodes := getNodes()
	key := "redLock"
	redLock, err := redislock.NewRedLock(key, nodes, redislock.WithRedLockExpire(10*time.Second), redislock.WithSingleNodeTimeout(100*time.Millisecond))
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		//lock1尝试获取锁
		if success, err := redLock.Lock(); success && err == nil {
			fmt.Println("go redLock get..")
			time.Sleep(4 * time.Second)
			redLock.Unlock()
		}
		wg.Done()
	}()

	//lock2尝试获取锁
	if success, err := redLock.Lock(); success && err == nil {
		fmt.Println("redLock get...")
		time.Sleep(7 * time.Second)
		redLock.Unlock()
	}
	wg.Wait()
}

func getNodes() []*redislock.SingleNode {
	//三个节点
	addr1 := "127.0.0.1:10000"
	passwd1 := "okredis"

	addr2 := "127.0.0.1:10001"
	passwd2 := "okredis"

	addr3 := "127.0.0.1:10002"
	passwd3 := "okredis"

	return []*redislock.SingleNode{
		{
			Address:  addr1,
			Password: passwd1,
		},
		{
			Address:  addr2,
			Password: passwd2,
		},
		{
			Address:  addr3,
			Password: passwd3,
		},
	}
}
