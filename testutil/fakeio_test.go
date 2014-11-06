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

func doTestRead(t *testing.T, r *FakeIO, resp []Resp) {
	p := make([]byte, 10)
	for i := range resp {
		t0 := time.Now()
		n, err := r.Read(p)
		t1 := time.Now()
		if n != resp[i].n {
			t.Fatalf("%d: n(%d) != %d", i, n, resp[i].n)
		}
		if err != resp[i].err {
			t.Fatalf("%d: Bad err: %s", i, err)
		}
		if !bytes.Equal(p[:n], resp[i].data) {
			t.Fatalf("%d: Data not equal!", i)
		}
		if t0.Add(r.Delay).After(t1) {
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
	r := NewFakeIO()
	r.Limit = 2
	r.ErrAfter = 9
	r.ErrEvery = 2
	r.Delay = 200 * time.Millisecond
	r.FillBytes(data)
	doTestRead(t, r, resp)
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
	r := NewFakeIO()
	r.Limit = 2
	r.ErrAfter = 0
	r.ErrEvery = 2
	r.Delay = 200 * time.Millisecond
	r.FillBytes(data)
	doTestRead(t, r, resp)
}

func TestFakeIOReadClose(t *testing.T) {
	r := NewFakeIO()
	r.Delay = 2 * time.Second
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

func doTestWrite(t *testing.T, w *FakeIO, data []byte, limit int) {
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
		if w.ErrAfter != 0 && i+1 > w.ErrAfter {
			if err != ErrPermanent {
				t.Fatalf("%d: Err not ErrPerm: %s", i, err)
			}
			return
		} else if w.ErrEvery != 0 &&
			(i+1)%w.ErrEvery == 0 &&
			err != ErrTemporary {
			t.Fatalf("%d: Err not ErrTemp: %s", i, err)
		}
		if t0.Add(w.Delay).After(t1) {
			t.Fatalf("%d: Short delay", i)
		}
		nn += n
	}
	if !bytes.Equal(w.Bytes(), data) {
		t.Fatal("Data not equal!")
	}
}

func TestFakeIOWrite0(t *testing.T) {
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	w := NewFakeIO()
	w.ErrAfter = 0
	w.ErrEvery = 2
	w.Delay = 200 * time.Millisecond
	doTestWrite(t, w, data, 2)
}

func TestFakeIOWrite1(t *testing.T) {
	data := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	w := NewFakeIO()
	w.ErrAfter = 4
	w.ErrEvery = 2
	w.Delay = 200 * time.Millisecond
	doTestWrite(t, w, data, 2)
}
