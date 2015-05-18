package main

import (
	"time"
)

func main() {
	i := 1
	skip := false

	fizz := time.Tick(3 * time.Second)
	time.Sleep(10 * time.Millisecond)
	buzz := time.Tick(5 * time.Second)
	time.Sleep(10 * time.Millisecond)
	num := time.Tick(1 * time.Second)

	for {
		select {
		case <-fizz:
			print("Fizz")
			skip = true
		case <-buzz:
			print("Buzz")
			skip = true
		case <-num:
			if !skip {
				print(i)
			}
			print("\n")
			skip = false
			i++
		}

		if i > 100 {
			return
		}
	}
}
