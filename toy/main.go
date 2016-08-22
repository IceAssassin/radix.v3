package main

import (
	"flag"
	"fmt"
	"github.com/Tsiannian/radix.v3/redis"
	"strings"
	"sync"
	"time"
)

func UnixTimeNow() float64 {
	now := time.Now().UnixNano()
	return float64(now) / 1000.0 / 1000.0 / 1000.0
}

var begin float64 = 0.0
var count int = 0
var countLock sync.Mutex = sync.Mutex{}

func AddCount(check float64) bool {
	now := UnixTimeNow()
	count++
	if begin < now-check {
		fmt.Println("ops", count, "in", check, "seconds")
		begin = now
		count = 0
		return true
	}
	return false
}

func main() {
	fmt.Println("main function begin")
	cmd := flag.String("cmd", "PING", "test redis command. ex.PING")

	flag.Parse()

	client, err := redis.DialTimeout("tcp", "127.0.0.1:6379", 10*time.Second)

	if err != nil {
		fmt.Println(err)
		return
	}

	cmdList := strings.Split(*cmd, " ")
	redisCmd := cmdList[0]
	redisArgs := make([]interface{}, len(cmdList)-1)
	for k, v := range cmdList[1:] {
		redisArgs[k] = v
	}

	time.AfterFunc(10*time.Second, func() { go client.Close(nil) })

	bench := func() {
		for true {
			info := client.FCmd(redisCmd, redisArgs...)
			if AddCount(1.0) {
				fmt.Printf("last cmd resp %v\n", info.GetResp())
			}
		}
	}

	bench()
}
