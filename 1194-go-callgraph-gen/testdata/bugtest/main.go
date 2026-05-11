package main

import "fmt"

type Counter struct {
	count int
}

func (c *Counter) Add(n int) {
	c.count += n
}

func (c *Counter) Get() int {
	return c.count
}

func helper() {
	fmt.Println("hi")
}

func baz() {
	c := &Counter{}
	c.Add(1)
	c.Get()
}

func main() {
	baz()
	helper()
}
