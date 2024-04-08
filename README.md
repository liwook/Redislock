# Redislock
基于go-redis实现的redis分布式锁
+ 支持阻塞模式获取锁,轮询等待锁   
+ 支持看门狗模式,可以对锁过期进行自动续期  
+ 支持红锁
     
go-redis连接使用redis  
  ```go​
func NewClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",    //自己的redis实例的ip和port
		Password: "",    //密码，有设置的话，就需要填写
	})
}
 
func main() {
	client := NewClient()
	defer client.Close()
 
	val, _ := client.Ping().Result()    //测试ping
	fmt.Println(val)
}
```
[Go语言实现Redis分布式锁](https://blog.csdn.net/m0_57408211/article/details/137335376)  
[Go语言实现Redis分布式锁2](https://blog.csdn.net/m0_57408211/article/details/137347252)
