package chanio

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/npat-efault/varhacks/pool"
	"github.com/npat-efault/varhacks/testutil"
)

func doTestRx(t *testing.T, data []byte, errEvery int, pl Pool) {
	r := testutil.NewFakeIO()
	r.ErrEvery = 2
	r.Delay = 200 * time.Millisecond
	r.FillBytes(data)
	rx := NewRx(r, 2, pl)
	var nn int
	for i := 1; ; i++ {
		p := <-rx.Buf()
		if i%r.ErrEvery == 0 {
			if p.Err != testutil.ErrTemporary {
				t.Fatalf("%d: Bad Error: %s", i, p.Err)
			}
		} else {
			if p.Err != nil {
				t.Fatalf("%d: Bad Error: %s", i, p.Err)
			}
			n := len(p.Data)
			if !bytes.Equal(data[nn:nn+n], p.Data) {
				t.Fatalf("%d: Bad data: %v", i, p.Data)
			}
			nn += n
			if nn == len(data) {
				break
			}
		}
	}
	p := <-rx.Buf()
	if p.Err != io.EOF {
		t.Fatal("No EOF:", p.Err)
	}
	err := rx.Close()
	if err != nil {
		t.Fatal("Close:", err)
	}
	// After Rx.Close channel rx.quit must be closed and channel
	// rx.Pck must be nil
	select {
	case <-rx.quit:
	default:
		t.Fatal("Channel rx.quit not closed")
	}
	if rx.Buf() != nil {
		t.Fatal("Channel rx.Pck() not nil")
	}
}

func TestRx(t *testing.T) {
	p := pool.NewByteSlice(10, nil)
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	doTestRx(t, data, 0, p)
	doTestRx(t, data, 2, nil)
	doTestRx(t, data, 4, p)
}

func doTestTx(t *testing.T, data []byte, sz int, errEvery int, pl Pool) {
	w := testutil.NewFakeIO()
	w.ErrEvery = errEvery
	w.Delay = 50 * time.Millisecond
	tx := NewTx(w, pl)
	var nn, i int
	for {
		i++
		if nn+sz > len(data) {
			sz = len(data) - nn
		}
		if sz == 0 {
			break
		}
		tx.Data() <- data[nn : nn+sz]
		r := <-tx.Res()
		if errEvery != 0 && (i%errEvery) == 0 {
			if r.Err != testutil.ErrTemporary {
				t.Fatalf("%d: Bad Error: %v", i, r.Err)
			}
		} else {
			if r.Err != nil {
				t.Fatalf("%d: Unexp Error: %v", i, r.Err)
			}
		}
		nn += sz
	}
	err := tx.Close()
	if err != nil {
		t.Fatal("Close:", err)
	}
	// After Tx.Close channel tx.quit must be closed and channel
	// tx.Pck must be nil
	select {
	case <-tx.quit:
	default:
		t.Fatal("Channel tx.quit not closed")
	}
	if tx.Data() != nil {
		t.Fatal("Channel tx.Pck() not nil")
	}
}

func TestTx(t *testing.T) {
	p := pool.NewByteSlice(10, nil)
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	doTestTx(t, data, 2, 0, p)
	doTestTx(t, data, 2, 2, p)
	doTestTx(t, data, 2, 3, p)
}

func TestRxTx(t *testing.T) {
	r := testutil.NewFakeIO()
	r.Limit = 4
	r.ErrEvery = 5
	r.Delay = 10 * time.Millisecond
	b := make([]byte, 1024)
	for i := 0; i < 1024; i++ {
		b[i] = byte(i)
	}
	r.FillBytes(b)
	rx := NewRx(r, 10, nil)

	w := testutil.NewFakeIO()
	w.ErrEvery = 3
	w.Delay = 8 * time.Millisecond
	tx := NewTx(w, nil)

	var p Buffer
	var rxp <-chan Buffer
	var txp chan<- []byte
	// arrange to receive first packet
	rxp = rx.Buf()
	txp = nil
loop:
	for {
		select {
		case p = <-rxp:
			if p.Err != nil {
				if p.Err == testutil.ErrTemporary {
					// retry receive
					break
				} else if p.Err == io.EOF {
					break loop
				} else {
					t.Fatal("Bad Error:", p.Err)
				}
			}
			// transmit packet
			rxp = nil
			txp = tx.Data()
		case txp <- p.Data:
		case r := <-tx.Res():
			if r.Err != nil {
				if r.Err != testutil.ErrTemporary {
					t.Fatal("Bad Error:", r.Err)
				}
				// retransmit
				break
			}
			// receive next packet
			rxp = rx.Buf()
			txp = nil
		}
	}
	rx.Close()
	tx.Close()
	if !bytes.Equal(b, w.Bytes()) {
		t.Fatal("Bad data!")
	}
}

// In TestRxTxQueued, the receiver and the transmitter are decoupled
// by a queue. Pearhaps a slightly cleaner implementation would be to
// have them as separate goroutines... There's no real need for that,
// though.

func TestRxTxQueued(t *testing.T) {
	r := testutil.NewFakeIO()
	r.Limit = 4
	r.ErrEvery = 5
	r.Delay = 20 * time.Millisecond
	b := make([]byte, 1024)
	for i := 0; i < 1024; i++ {
		b[i] = byte(i)
	}
	r.FillBytes(b)
	rx := NewRx(r, 10, nil)

	w := testutil.NewFakeIO()
	w.ErrEvery = 3
	w.Delay = 50 * time.Millisecond
	tx := NewTx(w, nil)

	q := make(chan []byte, 5)

	var p Buffer
	var rp []byte
	var rxp <-chan Buffer
	var txp chan<- []byte
	var q_in chan<- []byte
	var q_out <-chan []byte
	// Rx: arrange to receive first packet
	rxp = rx.Buf()
	q_in = nil
	// Tx: arrange to dequeue first packet
	q_out = q
	txp = nil

	var ok bool
loop:
	for {
		select {
		case p = <-rxp:
			// receive packet
			if p.Err != nil {
				if p.Err == testutil.ErrTemporary {
					// retry receive
					break
				} else if p.Err == io.EOF {
					close(q)
					// stop the receiver
					rxp = nil
					err := rx.Close()
					if err != nil {
						t.Fatal("rx.Close:", err)
					}
					break
				} else {
					t.Fatal("Bad Error:", p.Err)
				}
			}
			// enque packet
			rxp = nil
			q_in = q
		case q_in <- p.Data:
			// prepare to receive next
			rxp = rx.Buf()
			q_in = nil

		case rp, ok = <-q_out:
			// dequeue packet
			if !ok {
				// that's it, we' re done
				break loop
			}
			// prepare to transmit
			txp = tx.Data()
			q_out = nil
		case txp <- rp:
			// transmit
		case r := <-tx.Res():
			// get transmition result
			if r.Err != nil {
				if r.Err != testutil.ErrTemporary {
					t.Fatal("Bad Error:", r.Err)
				}
				// re-transmit
				break
			}
			// prepare to dequeue next packet
			txp = nil
			q_out = q
		}
	}
	err := tx.Close()
	if err != nil {
		t.Fatal("tx.Close:", err)
	}
	if !bytes.Equal(b, w.Bytes()) {
		t.Fatal("Bad data!")
	}
}
