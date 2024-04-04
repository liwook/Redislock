package main

import (
	"fmt"
	"redislock"

	"sync"
	"time"

	"github.com/go-redis/redis"
)

func NewClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "39.108.70.103:6379",
		Password: "wook1847",
	})
}

func main() {
	client := NewClient()
	defer client.Close()

	val, _ := client.Ping().Result()
	fmt.Println(val)

	key := "mylock"
	lock1 := redislock.NewRedisLock(client, key)
	lock2 := redislock.NewRedisLock(client, key)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		//lock1尝试获取锁
		if success, err := lock1.Lock(); success && err == nil {
			fmt.Println("go lock1 get..")
			time.Sleep(4 * time.Second)
			lock1.Unlock()
		}
		wg.Done()
	}()

	//lock2尝试获取锁
	if success, err := lock2.Lock(); success && err == nil {
		fmt.Println("lock2 get...")
		time.Sleep(7 * time.Second)
		lock2.Unlock()
	}
	wg.Wait()
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
