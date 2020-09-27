package main

import (
	"bufio"
	"log"
	"os"
)

var c chan int = make(chan int)

func a() {

	log.Printf("hi a\n")
	x := <-c
	log.Printf("a %v\n", x)

}

func b() {
	log.Printf("hi b\n")
	x := <-c
	log.Printf("b %v\n", x)
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	go a()
	go b()
	reader.ReadString('\n')
	c <- 1
	reader.ReadString('\n')
	c <- 2
	reader.ReadString('\n')
}
