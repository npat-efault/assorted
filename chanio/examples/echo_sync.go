package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/npat-efault/varhacks/chanio"
)

func Serve(c net.Conn, quit <-chan struct{}, fail chan<- struct{}) {
	var p chanio.Packet
	var rxp <-chan chanio.Packet
	var txp chan<- []byte
	var f chan<- struct{}
	var pool sync.Pool
	rx := chanio.NewRx(c, 4, &pool)
	tx := chanio.NewTx(c, &pool)
	rxp = rx.Pck()
	txp = nil
	f = nil
	var err error
	for {
		select {
		case p = <-rxp:
			if p.Err != nil {
				fmt.Println("Rx Error:", p.Err)
				err = rx.Close()
				fmt.Println("rx.Close:", err)
				rxp = nil
				err = tx.Close()
				fmt.Println("tx.Close:", err)
				txp = nil
				f = fail
				fmt.Println("Closed")
			} else {
				fmt.Println("Msg:", p.Pck)
				rxp = nil
				txp = tx.Pck()
			}
		case txp <- p.Pck:
			rxp = rx.Pck()
			txp = nil
		case res := <-tx.Err():
			fmt.Println("Tx Error:", res.Err)
			rx.Close()
			rxp = nil
			tx.Close()
			txp = nil
			f = fail
			fmt.Println("Closed")
		case f <- struct{}{}:
		case <-quit:
			err = tx.Drain()
			fmt.Println("tx.Drain:", err)
			err = rx.Close()
			fmt.Println("rx.Close:", err)
			fmt.Println("Quit")
			return
		}
	}
}

func main() {
	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Println("Listen:", err)
		os.Exit(1)
	}
	c, err := l.Accept()
	if err != nil {
		fmt.Println("Accept:", err)
		os.Exit(1)
	}
	quit := make(chan struct{})
	fail := make(chan struct{})
	go Serve(c, quit, fail)

	select {
	case <-fail:
		fmt.Println("Failure Report!")
		quit <- struct{}{}
	case <-time.After(15 * time.Second):
		fmt.Println("Quiting!")
		quit <- struct{}{}
	}

	time.Sleep(2 * time.Second)
	panic("Stacks!")
}
