package main

import (
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

const (
	maxRetries = 10
)

func main() {

	redisUrl := os.Getenv("REDIS_FLY_URL")
	u, err := url.Parse(redisUrl)

	if err != nil {
		panic(fmt.Sprintf("Invalid REDIS_URL: %s", err))
	}

	password := ""
	if u.User != nil {
		password, _ = u.User.Password()
	}

	addr := u.Host + ":6379"

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < maxRetries; j++ {
				client := redis.NewClient(&redis.Options{
					Addr:     addr,
					Password: password,
				})
				defer client.Close()

				_, err := client.Ping().Result()
				if err != nil {
					fmt.Printf("Error connecting to Redis server: %v\n", err)
					time.Sleep(time.Second)
					continue
				}

				fmt.Printf("Connection %d blocking on BLPOP...\n", id)
				val, err := client.BLPop(time.Duration(time.Duration.Seconds(5)), "test").Result()
				fmt.Println(val)
				if err != nil {
					fmt.Printf("Error blocking on Redis list: %v\n", err)
					time.Sleep(time.Second)
					continue
				}
				break
			}
		}(i)
	}
	wg.Wait()
}
