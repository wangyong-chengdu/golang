package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var balance int32
	balance = int32(0)
	done := make(chan bool)
	count := 1000
	//var lock sync.Mutex
	for i := 0; i < count; i++ {
		//go transfer(&balance, 1, done, &lock)
		//go transferCas(&balance, 1, done)
		go transferFaa(&balance, 1, done)
	}

	for i := 0; i < count; i++ {
		<-done
	}
	fmt.Printf("balance = %d \n", balance)
}

func transfer(balance *int32, amount int, done chan bool, lock *sync.Mutex) {
	lock.Lock()
	*balance = *balance + int32(amount)
	lock.Unlock()
	done <- true
}

func transferCas(balance *int32, amount int, done chan bool) {
	for {
		old := atomic.LoadInt32(balance)
		new := old + int32(amount)
		if atomic.CompareAndSwapInt32(balance, old, new) {
			break
		}
	}
	done <- true
}

func transferFaa(balance *int32, amount int, done chan bool) {
	atomic.AddInt32(balance, int32(amount))
	done <- true
}
