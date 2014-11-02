package testutil

import (
	"bytes"
	"io"
	"testing"
	"time"
)

type Resp struct {
	n    int
	err  error
	data []byte
}

func doTestRead(t *testing.T, r *FakeIO, resp []Resp, dly time.Duration) {
	p := make([]byte, 10)
	for i := range resp {
		t0 := time.Now()
		n, err := r.Read(p)
		t1 := time.Now()
		if n != resp[i].n {
			t.Fatalf("%d: n(%d) != %d", i, n, resp[i].n)
		}
		if err != resp[i].err {
			t.Fatalf("%d: Bad err: %v", i, err)
		}
		if !bytes.Equal(p[:n], resp[i].data) {
			t.Fatalf("%d: Data not equal!", i)
		}
		if t0.Add(dly).After(t1) {
			t.Fatalf("%d: Short delay", i)
		}
	}
}

func TestFakeIORead0(t *testing.T) {
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	resp := []Resp{
		{n: 2, err: nil, data: []byte{0, 1}},
		{n: 0, err: ErrTemporary, data: []byte{}},
		{n: 2, err: nil, data: []byte{2, 3}},
		{n: 0, err: ErrTemporary, data: []byte{}},
		{n: 2, err: nil, data: []byte{4, 5}},
		{n: 0, err: ErrTemporary, data: []byte{}},
		{n: 2, err: nil, data: []byte{6, 7}},
		{n: 0, err: ErrTemporary, data: []byte{}},
		{n: 2, err: nil, data: []byte{8, 9}},
		{n: 0, err: ErrPermanent, data: []byte{}},
		{n: 0, err: ErrPermanent, data: []byte{}},
	}
	dly := 200 * time.Millisecond
	r := NewFakeIO(
		2,   // limit
		9,   // errAfter
		2,   // errEvery
		dly) // delay
	r.FillBytes(data)
	doTestRead(t, r, resp, dly)
}

func TestFakeIORead1(t *testing.T) {
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	resp := []Resp{
		{n: 2, err: nil, data: []byte{0, 1}},
		{n: 0, err: ErrTemporary, data: []byte{}},
		{n: 2, err: nil, data: []byte{2, 3}},
		{n: 0, err: ErrTemporary, data: []byte{}},
		{n: 2, err: nil, data: []byte{4, 5}},
		{n: 0, err: ErrTemporary, data: []byte{}},
		{n: 2, err: nil, data: []byte{6, 7}},
		{n: 0, err: ErrTemporary, data: []byte{}},
		{n: 2, err: nil, data: []byte{8, 9}},
		{n: 0, err: ErrTemporary, data: []byte{}},
		{n: 1, err: nil, data: []byte{10}},
		{n: 0, err: io.EOF, data: []byte{}},
		{n: 0, err: io.EOF, data: []byte{}},
		{n: 0, err: io.EOF, data: []byte{}},
	}
	dly := 200 * time.Millisecond
	r := NewFakeIO(
		2,   // limit
		0,   // errAfter
		2,   // errEvery
		dly) // delay
	r.FillBytes(data)
	doTestRead(t, r, resp, dly)
}

func TestFakeIOReadClose(t *testing.T) {
	r := NewFakeIO(
		0,             // limit
		0,             // errAfter
		0,             // errEvery
		2*time.Second) // delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		r.Close()
	}()
	p := make([]byte, 1)
	n, err := r.Read(p)
	if n != 0 || err != ErrClosed {
		t.Fatal("FakeIO not closed:", n, err)
	}
}

func doTestWrite(t *testing.T, w *FakeIO, data []byte,
	limit, errAfter, errEvery int, dly time.Duration) {

	for i, nn := 0, 0; nn < len(data); i++ {
		var n int
		var err error
		t0 := time.Now()
		if len(data)-nn >= limit {
			n, err = w.Write(data[nn : nn+limit])
		} else {
			n, err = w.Write(data[nn:])
		}
		t1 := time.Now()
		if errAfter != 0 && i+1 > errAfter {
			if err != ErrPermanent {
				t.Fatalf("%d: Err not ErrPerm: %v", i, err)
			}
			return
		} else if errEvery != 0 &&
			(i+1)%errEvery == 0 &&
			err != ErrTemporary {
			t.Fatalf("%d: Err not ErrTemp: %v", i, err)
		}
		if t0.Add(dly).After(t1) {
			t.Fatalf("%d: Short delay", i)
		}
		nn += n
	}
	if !bytes.Equal(w.Bytes(), data) {
		t.Fatalf("Data not equal!")
	}
}

func TestFakeIOWrite0(t *testing.T) {
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	limit := 2
	errAfter := 0
	errEvery := 2
	dly := 200 * time.Millisecond
	w := NewFakeIO(
		0,
		errAfter,
		errEvery,
		dly)
	doTestWrite(t, w, data, limit, errAfter, errEvery, dly)
}

func TestFakeIOWrite1(t *testing.T) {
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	limit := 2
	errAfter := 4
	errEvery := 2
	dly := 200 * time.Millisecond
	w := NewFakeIO(
		0,
		errAfter,
		errEvery,
		dly)
	doTestWrite(t, w, data, limit, errAfter, errEvery, dly)
}
