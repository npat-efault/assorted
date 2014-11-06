// Package chanio provides a channel interface for io.Readers,
// io.Writers, and net.Listeners. It implements three types: Rx
// (receiver), Tx (transmitter), and Lx (Listener). See example in:
//
package chanio

import (
	"errors"
	"io"
	"net"
)

var ErrClosed = errors.New("Rx/Tx/Lx already closed")

type Pool interface {
	Get() []byte
	Put([]byte)
}

// Buffer is the type received (by the user) from the Rx.Buf()
// channel.
type Buffer struct {
	Data []byte
	Err  error
}

// Rx provides a channel interface for reading (receiving) data from
// an io.ReadCloser.
type Rx struct {
	r        io.ReadCloser
	maxPckSz int
	pool     Pool
	pbuf     []byte
	cbuf     chan Buffer
	quit     chan struct{}
}

// NewRx creates and returns an Rx receiver. It spawns a goroutine
// that reads from the supplied io.ReadCloser and makes the data read
// (and / or any errors detected) available through a channel of type
// "chan Buffer". The Read method of the io.ReaderCloser is called by
// the Rx with buffers of length == "maxPckSz". If the "pool" argument
// is not nil, its Get() method is called to supply the buffers. The
// buffers returned by the pool *must* have capacity >= "maxPckSz". If
// "pool" is nil, or if pool.Get() returns nil, new buffers are
// allocated by the Rx.
func NewRx(r io.ReadCloser, maxPckSz int, pool Pool) *Rx {
	rx := &Rx{}
	rx.r = r
	rx.maxPckSz = maxPckSz
	rx.pool = pool
	if pool == nil {
		rx.pbuf = make([]byte, maxPckSz)
	}
	rx.cbuf = make(chan Buffer)
	rx.quit = make(chan struct{})
	go rx.run()
	return rx
}

// Buf returns the channel where reader-data (and any detected errors)
// can be received from.
func (rx *Rx) Buf() <-chan Buffer {
	return rx.cbuf
}

// Close terminates the operation of the receiver and releases the
// respective goroutine. Subsequent reads from the Rx.Buf() channel
// will always block. Close can be called multiple times (it will
// return ErrClosed after the first). It is *not* safe to call Close
// concurently from multiple goroutines.
func (rx *Rx) Close() error {
	if rx.cbuf == nil {
		return ErrClosed
	}
	err := rx.r.Close()
	rx.quit <- struct{}{}
	close(rx.quit) // Concurent calls to rx.Close may panic
	rx.cbuf = nil
	return err
}

func (rx *Rx) run() {
	var err error
	var p []byte
	for {
		if rx.pool != nil {
			p = rx.pool.Get()
			if p != nil {
				p = p[:rx.maxPckSz]
			} else {
				p = make([]byte, rx.maxPckSz)
			}
			var n int
			n, err = rx.r.Read(p)
			p = p[:n]
		} else {
			var n int
			n, err = rx.r.Read(rx.pbuf)
			p = make([]byte, n)
			copy(p, rx.pbuf)
		}
		select {
		case <-rx.quit:
			return
		case rx.cbuf <- Buffer{p, err}:
		}
	}
}

// Result is the type received (by the user) from the Tx.Err()
// channel. It is sent by Tx to indicate an error durring the
// transmission of the last buffer. N is used by the writer to
// indicate the number of bytes transmitted before the error (if
// applicable) and Err to indicate the error.
type Result struct {
	N   int
	Err error
}

// Tx provides a channel interface for writing (sending) data to an
// io.WriteCloser.
type Tx struct {
	w     io.WriteCloser
	pool  Pool
	cdata chan []byte
	res   chan Result
	quit  chan struct{}
}

// NewTx creates and returns a Tx transmitter. It spawns a goroutine
// that writes to the supplied io.WriteCloser data send by the user on
// the Tx.Data() channel (of type "chan<- []byte"). If the "pool"
// argument is not nil, after the data are transmitter the buffer is
// returned to the pool by calling pool.Put().
func NewTx(w io.WriteCloser, pool Pool) *Tx {
	tx := &Tx{}
	tx.w = w
	tx.pool = pool
	tx.cdata = make(chan []byte)
	tx.res = make(chan Result)
	tx.quit = make(chan struct{})
	go tx.run()
	return tx
}

// Data returns the channel where data can be sent to.
func (tx *Tx) Data() chan<- []byte {
	return tx.cdata
}

// Res returns the channel where the user receives success or error
// reports (results) for the transmitted data. After a buffer is
// transmitted, Rx.Tx sends a Result structure on this channel
// reporting whether the transmission was succesful, or not. Rx.Tx
// will not accept new data until the user has received this result.
func (tx *Tx) Res() <-chan Result {
	return tx.res
}

// Close immediately terminates the operation of the transmitter and
// releases the respective goroutine. Subsequent writes to the
// Tx.Data() channel or reads from the Tx.Res() channel will always
// block. Close can be called multiple times (it will return ErrClosed
// after the first). It is *not* safe to call Close concurently from
// multiple goroutines.
func (tx *Tx) Close() error {
	if tx.cdata == nil {
		return ErrClosed
	}
	err := tx.w.Close()
	tx.quit <- struct{}{}
	close(tx.quit) // Concur. calls to tx.Close/Drain may panic
	tx.cdata = nil
	tx.res = nil
	return err
}

func (tx *Tx) run() {
	for {
		var err error
		var n int

		// wait for data
		select {
		case p := <-tx.cdata:
			n, err = tx.w.Write(p)
			if tx.pool != nil {
				tx.pool.Put(p)
			}
		case <-tx.quit:
			return
		}
		// send back result
		select {
		case tx.res <- Result{n, err}:
		case <-tx.quit:
			return
		}
	}
}

// Connection is the type received (by the user) from the Lx.Conn()
// channel. It is sent by Lx to indicate that a new connection is
// available, or than an error was reported by the listener.
type Connection struct {
	Conn net.Conn
	Err  error
}

// Lx provides a channel interface for accepting network connections.
type Lx struct {
	l     net.Listener
	cconn chan Connection
	quit  chan struct{}
}

// NewLx creates and returns a new Lx listener. It spawns a goroutine
// that uses the supplied net.Listener to accept connections and sends
// them through the Lx.Conn() channel to the user.
func NewLx(l net.Listener) *Lx {
	lx := &Lx{}
	lx.l = l
	lx.cconn = make(chan Connection)
	lx.quit = make(chan struct{})
	go lx.run()
	return lx
}

// Conn returns the channel where connections can be received from.
func (lx *Lx) Conn() <-chan Connection {
	return lx.cconn
}

// Close terminates the operation of the listener and releases the
// respective goroutine. Subsequent reads from the Lx.Conn() channel
// will always block. Close can be called multiple times (it will
// return ErrClosed after the first). It is *not* safe to call Close
// concurently from multiple goroutines.
func (lx *Lx) Close() error {
	if lx.cconn == nil {
		return ErrClosed
	}
	err := lx.l.Close()
	lx.quit <- struct{}{}
	close(lx.quit) // Concurent calls to lx.Close may panic
	lx.cconn = nil
	return err
}

func (lx *Lx) run() {
	var err error
	var c net.Conn
	for {
		c, err = lx.l.Accept()
		select {
		case <-lx.quit:
			return
		case lx.cconn <- Connection{c, err}:
		}
	}
}

// NOTE(npat): In order to make Close() concurency-safe (and allow the
// same chanio.Rx/Lx/Tx to be used, safely, by multiple goroutines),
// all that's required is to protect the Close() methods and the
// Buf()/Data()/Res()/Conn() accessors with a mutex.
