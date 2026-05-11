package main

import (
	"fmt"
	"time"
)

type Counter struct {
	count int
}

func (c *Counter) Add(n int) {
	c.count += n
}

func (c *Counter) Get() int {
	return c.count
}

func init() {
	fmt.Println("init called")
}

func foo() {
	bar()
}

func bar() {
	baz()
}

func baz() {
	c := &Counter{}
	c.Add(1)
	fmt.Println(c.Get())
}

func worker() {
	time.Sleep(time.Second)
	fmt.Println("worker done")
}

func main() {
	foo()

	defer func() {
		fmt.Println("deferred")
	}()

	go func() {
		worker()
	}()

	go worker()

	time.Sleep(2 * time.Second)
}
