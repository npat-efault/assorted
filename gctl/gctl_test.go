// Copyright (c) 2015, Nick Patavalis (npat@efault.net).
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can
// be found in the LICENSE file.

package gctl

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestRaceGcxWait(t *testing.T) {
	// data race on Gcx.kill
	var gcx Gcx
	go gcx.Go(func() error { return nil })
	time.Sleep(10 * time.Millisecond)
	gcx.Wait()
}

func TestRaceGcxChKill(t *testing.T) {
	// data race on Gcx.kill
	var gcx Gcx
	go gcx.Go(func() error { return nil })
	time.Sleep(10 * time.Millisecond)
	gcx.ChKill()
}

func TestRaceGcxKill(t *testing.T) {
	// data race on Gcx.kill
	var gcx Gcx
	go gcx.Go(func() error { return nil })
	time.Sleep(10 * time.Millisecond)
	gcx.Kill()
}

func TestRaceGcxKillSignaled(t *testing.T) {
	// data race on Gcx.signaled
	var gcx Gcx
	go gcx.Go(func() error { return nil })
	time.Sleep(10 * time.Millisecond)
	go gcx.Kill()
	time.Sleep(10 * time.Millisecond)
	gcx.Kill()
}

func TestGcxChKillPanic(t *testing.T) {
	defer func() {
		x := recover()
		s, ok := x.(string)
		if !ok || !strings.HasPrefix(s, "Gcx.") {
			panic(x)
		}
	}()

	var gcx Gcx
	gcx.ChKill()
	t.Fatal("ChKill should panic")
}

func TestGcxGoErrStarted(t *testing.T) {
	var gcx Gcx
	err := gcx.Go(func() error { return nil })
	if err != nil {
		t.Fatalf("gcx.Go: %v != nil", err)
	}
	err = gcx.Wait()
	if err != nil {
		t.Fatalf("Wait: %v != nil", err)
	}
	err = gcx.Go(func() error { return nil })
	if err != ErrStarted {
		t.Fatalf("gcx.Go: %v != ErrStarted", err)
	}
	gcx = GcxZero
	err = gcx.Go(func() error { return nil })
	if err != nil {
		t.Fatalf("gcx.Go: %v != nil", err)
	}

}

func TestSGcxGoMany(t *testing.T) {
	N := 42
	cc := make(chan *Gcx, 10000)
	fn := func(fn func() error) error {
		for i := 0; i < N; i++ {
			wms := time.Duration(rand.Intn(10) + 1)
			time.Sleep(wms * time.Millisecond)
			c := &Gcx{}
			c.Go(fn)
			cc <- c
		}
		return nil
	}
	fnX := func() error {
		wms := time.Duration(rand.Intn(100) + 1)
		time.Sleep(wms * time.Millisecond)
		return nil
	}
	fn1 := func() error {
		fn(fnX)
		return nil
	}
	fn0 := func() error {
		fn(fn1)
		return nil
	}

	fn0()

	n := 0
loop:
	for {
		select {
		case c := <-cc:
			e := c.Wait()
			if e != nil {
				t.Fatalf("xs for %d: %v", n, e)
			}
			n++
		case <-time.After(1 * time.Second):
			break loop
		}
	}
	if n != N+N*N {
		t.Fatalf("Waited for %d != %d goroutines", n, N+N*N)
	}
}

func TestKillWait(t *testing.T) {
	var c Gcx
	var ErrKilled = errors.New("Killed")
	if e := c.Kill(); e != ErrNotStarted {
		t.Fatalf("Gcx.Kill: %v", e)
	}
	if e := c.Wait(); e != ErrNotStarted {
		t.Fatalf("Gcx.Wait: %v", e)
	}
	e := c.Go(func() error {
		select {
		case <-c.ChKill():
			return ErrKilled
		case <-time.After(1 * time.Second):
			return nil
		}
	})
	if e != nil {
		t.Fatalf("Gcx.Go: %v", e)
	}
	// Kill it
	if e := c.Kill(); e != nil {
		t.Fatalf("Gcx.Kill: %v", e)
	}
	// Try killing a second time
	if e := c.KillWait(); e != ErrKilled {
		t.Fatalf("Gcx.KillWait: %v", e)
	}
	// And once more
	if e := c.Wait(); e != ErrKilled {
		t.Fatalf("Gcx.Wait: %v", e)
	}
}
