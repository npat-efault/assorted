package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"github.com/npat-efault/varhacks/chanio"
)

type Block interface {
	Id() int
	Fail() error
}

var AllBlocks = map[int]Block{}

type Echoer struct {
	c    net.Conn
	fail chan<- int
	q    chan []byte
	rcv  *Receiver
	trx  *Transmitter
}

func NewEchoer(c net.Conn, fail chan<- int) *Echoer {
	fmt.Println("Starting Echoer...")
	e := &Echoer{c: c, fail: fail}
	e.q = make(chan []byte, 5)
	e.rcv = NewReceiver(0, e)
	e.trx = NewTransmitter(1, e)
	fmt.Println("Echoer started")
	return e
}

func (e *Echoer) Stop() error {
	fmt.Println("Stopping Echoer...")
	e.rcv.Stop()
	e.trx.Stop()
	fmt.Println("Echoer stopped")
	return nil
}

type Receiver struct {
	id   int
	e    *Echoer
	rx   *chanio.Rx
	quit chan chan error
}

func NewReceiver(id int, e *Echoer) *Receiver {
	fmt.Println("Starting Receiver...")
	r := &Receiver{id: id, e: e}
	r.rx = chanio.NewRx(r.e.c, 4, nil)
	r.quit = make(chan chan error)
	AllBlocks[r.id] = r
	go r.run()
	fmt.Println("Receiver started.")
	return r
}

func (r *Receiver) Stop() error {
	fmt.Println("Stopping Receiver...")
	ce := make(chan error)
	r.quit <- ce
	err := <-ce
	fmt.Println("Receiver stopped")
	return err
}

func (r *Receiver) Id() int {
	return r.id
}

func (r *Receiver) Fail() error {
	return r.e.Stop()
}

func (r *Receiver) run() {
	var p chanio.Packet
	var rxp <-chan chanio.Packet = r.rx.Pck()
	var txp chan<- []byte = nil
	var f chan<- int = nil
	for {
		select {
		case p = <-rxp:
			if p.Err != nil {
				f = r.e.fail
				rxp = nil
				txp = nil
			} else {
				fmt.Println("Msg:", p.Pck)
				rxp = nil
				txp = r.e.q
			}
		case txp <- p.Pck:
			txp = nil
			rxp = r.rx.Pck()
		case f <- r.id:
			f = nil
		case ce := <-r.quit:
			err := r.rx.Close()
			ce <- err
			fmt.Println("Receiver goroutine exit")
			return
		}
	}
}

type Transmitter struct {
	id   int
	e    *Echoer
	tx   *chanio.Tx
	quit chan chan error
}

func NewTransmitter(id int, e *Echoer) *Transmitter {
	fmt.Println("Starting Transmitter...")
	t := &Transmitter{id: id, e: e}
	t.tx = chanio.NewTx(t.e.c, nil)
	t.quit = make(chan chan error)
	AllBlocks[t.id] = t
	go t.run()
	fmt.Println("Transmitter started.")
	return t
}

func (t *Transmitter) Stop() error {
	fmt.Println("Stopping Transmitter...")
	ce := make(chan error)
	t.quit <- ce
	err := <-ce
	fmt.Println("Transmitter stopped.")
	return err
}

func (t *Transmitter) Id() int {
	return t.id
}

func (t *Transmitter) Fail() error {
	return t.e.Stop()
}

func (t *Transmitter) run() {
	var p []byte
	var rxp <-chan []byte = t.e.q
	var txp chan<- []byte = nil
	var f chan<- int = nil
	for {
		select {
		case p = <-rxp:
			rxp = nil
			txp = t.tx.Pck()
		case txp <- p:
			txp = nil
			rxp = t.e.q
		case <-t.tx.Err():
			f = t.e.fail
			txp = nil
			rxp = nil
		case f <- t.id:
			f = nil
		case ce := <-t.quit:
			err := t.tx.Close()
			ce <- err
			fmt.Println("Transmitter goroutine exit")
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
	fail := make(chan int)
	e := NewEchoer(c, fail)
	select {
	case id := <-fail:
		fmt.Println("Failure Report from:", id)
		blk := AllBlocks[id]
		if blk != nil {
			fmt.Println("Calling Fail for block:", id)
			blk.Fail()
		} else {
			fmt.Println("Unknown block failure:", id)
		}
	case <-time.After(15 * time.Second):
		e.Stop()
	}

	time.Sleep(2 * time.Second)
	panic("Stacks!")
}
