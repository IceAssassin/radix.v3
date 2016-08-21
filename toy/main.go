package main

import (
	"fmt"
	"time"
	//"flag"
	"../redis"
)

func main() {
	fmt.Println("main function begin")
	//cmd := flag.String("cmd", "PING", "test redis command. ex.PING")
	//flag.Parse()

	client, err := redis.DialTimeout("tcp", "127.0.0.1:6379", 10*time.Second)

	if err != nil {
		fmt.Println(err)
		return
	}

	i := 1

	go func() {
		ret := client.ReadResp()
		fmt.Println(ret)
	}()

	for i > 0 {
		client.Cmd("INCRTEST")
		//fmt.Println(ret)
	}
}