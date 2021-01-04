package main

import (
	"fmt"
	"time"
)

func hello(msg string) {
	fmt.Println("Hello " + msg)
}

func main() {
	go hello("World")
	fmt.Println("Run in main")
	time.Sleep(100 * time.Millisecond)
}
