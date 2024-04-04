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
		Addr:     "127.0.0.1:6379", //redis实例的ip ,port
		Password: "",               //有设置密码就填写密码
	})
}

func main() {
	client := NewClient()
	defer client.Close()

	val, _ := client.Ping().Result()
	fmt.Println(val)

	key := "mylock"
	lock := redislock.NewRedisLock(client, key)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		//尝试获取锁
		if success, err := lock.Lock(); success && err == nil {
			fmt.Println("go lock get..")
			time.Sleep(4 * time.Second)
			lock.Unlock()
		}
		wg.Done()
	}()

	//尝试获取锁
	// time.Sleep(1 * time.Second)
	if success, err := lock.Lock(); success && err == nil {
		fmt.Println(" main lock get...")
		time.Sleep(7 * time.Second)
		lock.Unlock()
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
