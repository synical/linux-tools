package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func readNetBytes(c chan int, dev string, direction string) {
	for {
		b, _ := ioutil.ReadFile("/sys/class/net/" + dev + "/statistics/" + direction + "_bytes")
		bs := string(b)
		bs = strings.Replace(bs, "\n", "", -1)
		bi, _ := strconv.Atoi(bs)
		c <- bi
		time.Sleep(time.Second * 1)
	}
}

func main() {
	rxChan := make(chan int)
	txChan := make(chan int)
	go readNetBytes(rxChan, "eth0", "rx")
	go readNetBytes(txChan, "eth0", "tx")
	for {
		select {
		case rx := <-rxChan:
			fmt.Println(rx)
		case tx := <-txChan:
			fmt.Println(tx)
		}
	}
}
