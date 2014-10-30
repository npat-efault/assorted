// chanio provides a packetized channel interface for io.Readers and
// io.Writers (actually io.ReadClosers and io.WriteClosers). It
// implements two types: Rx (receiver) and Tx (transmitter).
//
package chanio

import (
	"errors"
	"io"
)

var ErrClosed = errors.New("Rx/Tx already closed")

type Pool interface {
	Get() interface{}
	Put(interface{})
}

// Packet is the type of data received from the Rx.Pck() channel.
type Packet struct {
	Pck []byte
	Err error
}

// Rx provides a packetized channel interface for reading (receiving)
// data from an io.ReadCloser.
type Rx struct {
	r        io.ReadCloser
	maxPckSz int
	pool     Pool
	pck      chan Packet
	quit     chan struct{}
	done     chan struct{}
}

// NewRx creates and returns an Rx receiver. It spawns a goroutine
// that reads from the supplied io.ReadCloser and makes the data read
// (and / or any errors detected) available through a channel of type
// "chan Packet". The Read method of the io.ReaderCloser is called by
// Rx with buffers of length == maxPckSz. If the pool argument is
// non-nil its Get() method is called to retrieve buffers. The buffers
// returned by the pool *must* have capacity >= maxPckSz. If pool ==
// nil or if poolGet() returns nil, new buffers are allocated.
func NewRx(r io.ReadCloser, maxPckSz int, pool Pool) *Rx {
	rx := &Rx{}
	rx.r = r
	rx.maxPckSz = maxPckSz
	rx.pool = pool
	rx.pck = make(chan Packet)
	rx.quit = make(chan struct{})
	rx.done = make(chan struct{})
	go rx.run()
	return rx
}

// Pck returns the channel where reader-data (and any detected errors)
// can be received from.
func (rx *Rx) Pck() <-chan Packet {
	return rx.pck
}

// Close terminates the operation of the receiver and releases the
// respective goroutine. Subsequent reads from the Rx.Pck() channel
// will always block. Close can be called multiple times (it will
// return ErrClosed after the first). It is *not* safe to call Close
// concurently from multiple goroutines.
func (rx *Rx) Close() error {
	if rx.pck == nil {
		return ErrClosed
	}
	err := rx.r.Close()
	rx.quit <- struct{}{}
	close(rx.quit) // Concurent calls to rx.Close will panic
	rx.pck = nil
	return err
}

func (rx *Rx) run() {
	var err error
	var p, p0 []byte
	if rx.pool == nil {
		p0 = make([]byte, rx.maxPckSz)
	}
	for {
		if rx.pool != nil {
			var n int
			if pp := rx.pool.Get(); pp != nil {
				p = pp.([]byte)[:rx.maxPckSz]
			} else {
				p = make([]byte, rx.maxPckSz)
			}
			n, err = rx.r.Read(p)
			p = p[:n]
		} else {
			var n int
			n, err = rx.r.Read(p0)
			p = make([]byte, n)
			copy(p, p0)
		}
		select {
		case <-rx.quit:
			return
		case rx.pck <- Packet{p, err}:
		}
	}
}

type Result struct {
	N   int
	Err error
}

type Tx struct {
	w    io.WriteCloser
	pool Pool
	pck  chan []byte
	err  chan Result
	quit chan struct{}
}

func NewTx(w io.WriteCloser, pool Pool) *Tx {
	tx := &Tx{}
	tx.w = w
	tx.pool = pool
	tx.pck = make(chan []byte)
	tx.err = make(chan Result)
	tx.quit = make(chan struct{})
	go tx.run()
	return tx
}

func (tx *Tx) Pck() chan<- []byte {
	return tx.pck
}

func (tx *Tx) Err() <-chan Result {
	return tx.err
}

func (tx *Tx) Close() error {
	if tx.pck == nil {
		return ErrClosed
	}
	err := tx.w.Close()
	tx.quit <- struct{}{}
	close(tx.quit) // Concur. calls to tx.Close/Drain will panic
	tx.pck = nil
	tx.err = nil
	return err
}

func (tx *Tx) Drain() error {
	if tx.pck == nil {
		return ErrClosed
	}
	tx.quit <- struct{}{}
	err := tx.w.Close()
	close(tx.quit) // Concur. calls to tx.Close/Drain will panic
	tx.pck = nil
	tx.err = nil
	return err
}

func (tx *Tx) run() {
	var err error
	var n int
	for {
		if err == nil {
			select {
			case p := <-tx.pck:
				n, err = tx.w.Write(p)
				if tx.pool != nil {
					tx.pool.Put(p)
				}
			case <-tx.quit:
				return
			}
		} else {
			select {
			case tx.err <- Result{n, err}:
				err = nil
			case <-tx.quit:
				return
			}
		}
	}
}
