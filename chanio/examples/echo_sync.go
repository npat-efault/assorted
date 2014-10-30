package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"github.com/npat-efault/varhacks/chanio"
)

// fail: Echo notifies main that is has failed (encountered error)
// quit: main writes to terminate Echo (even after fail)
// done: Echo notifies main that it is all done (about to exit)
func Echo(c net.Conn,
	fail chan<- struct{}, quit <-chan struct{}, done chan<- struct{}) {

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
				rxp = nil
				txp = nil
				f = fail
				fmt.Println("Failure")
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
			rxp = nil
			txp = nil
			f = fail
			fmt.Println("Failure")
		case f <- struct{}{}:
		case <-quit:
			err = tx.Drain()
			fmt.Println("tx.Drain:", err)
			err = rx.Close()
			fmt.Println("rx.Close:", err)
			fmt.Println("Quit")
			close(done)
			return
		}
	}
}

func Usage(cmd string) {
	fmt.Fprintf(os.Stderr, "Usage is: %s <local addr>\n", cmd)
}

func main() {
	if len(os.Args) != 2 {
		Usage(path.Base(os.Args[0]))
		os.Exit(1)
	}
	l, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Listen:", err)
		os.Exit(1)
	}
	c, err := l.Accept()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Accept:", err)
		os.Exit(1)
	}
	fail := make(chan struct{})
	quit := make(chan struct{})
	done := make(chan struct{})
	go Echo(c, fail, quit, done)

	select {
	case <-fail:
		fmt.Println("Failure Report!")
		quit <- struct{}{}
		<-done
	case <-time.After(15 * time.Second):
		fmt.Println("Quiting!")
		quit <- struct{}{}
		<-done
	}

	// time.Sleep(2 * time.Second)
	// panic("Stacks!")
}
