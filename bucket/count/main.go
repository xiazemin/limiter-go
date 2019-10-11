package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var lr LimitRate
	lr.SetRate(3)

	for i:=0;i<10;i++{
		wg.Add(1)
		go func(){
			if lr.Limit() {
				fmt.Println("Got it!")//显示3次Got it!
			}
			wg.Done()
		}()
	}
	wg.Wait()
}