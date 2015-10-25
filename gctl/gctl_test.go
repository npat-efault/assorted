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

func TestGroupUninit(t *testing.T) {
	var c *Gcx
	var xs error
	var g Group
	c, xs = g.Wait()
	if c != nil || xs != nil {
		t.Fatalf("g.Wait: c = %p, xs = %v", c, xs)
	}
	c, xs = g.Poll()
	if c != nil || xs != nil {
		t.Fatalf("g.Poll: c = %p, xs = %v", c, xs)
	}
	if n := g.Count(); n != 0 {
		t.Fatalf("g.Count: n = %d", n)
	}
}

func TestGroupSet(t *testing.T) {
	var g Group
	var c Gcx
	var x *Gcx
	var xs error

	c.Go(func() error { return nil })
	if e := c.SetGroup(&g); e != ErrStarted {
		t.Fatalf("Gcx.SetGroup: %v", e)
	}
	c.Wait()
	c = GcxZero

	if e := c.SetGroup(&g); e != nil {
		t.Fatalf("Gcx.SetGroup: %v", e)
	}
	if n := g.Count(); n != 0 {
		t.Fatalf("Group.Count: %d", n)
	}
	x, xs = g.Wait()
	if x != nil || xs != nil {
		t.Fatalf("g.Wait: x = %p, xs = %v", x, xs)
	}

	c.Go(func() error { return nil })
	if n := g.Count(); n != 1 {
		t.Fatalf("Group.Count: %d", n)
	}
	x, xs = g.Wait()
	if x != &c || xs != nil {
		t.Fatalf("g.Wait: x = %p, xs = %v", x, xs)
	}
	x, xs = g.Wait()
	if x != nil || xs != nil {
		t.Fatalf("g.Wait: x = %p, xs = %v", x, xs)
	}
	if n := g.Count(); n != 0 {
		t.Fatalf("Group.Count: %d", n)
	}
}

func TestGroupMany(t *testing.T) {
	N := 42
	g := &Group{}
	fn := func(fn func() error) error {
		for i := 0; i < N; i++ {
			wms := time.Duration(rand.Intn(10) + 1)
			time.Sleep(wms * time.Millisecond)
			c := &Gcx{}
			c.SetGroup(g)
			c.Go(fn)
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

	go fn0()

	time.Sleep(200 * time.Millisecond)
	n := 0

	// Grab a few
	for c, xs := g.Poll(); c != nil; c, xs = g.Poll() {
		if xs != nil {
			t.Fatalf("%d: %v", n, xs)
		}
		n++
	}
	//n1 := n

	// Wait for all to start
	time.Sleep(4 * time.Second)

	if cnt := g.Count(); cnt != (N+N*N)-n {
		t.Fatalf("g.Count %d != %d", cnt, (N+N*N)-n)
	}

	for c, xs := g.Wait(); c != nil; c, xs = g.Wait() {
		if xs != nil {
			t.Fatalf("%d: %v", n, xs)
		}
		n++
		if cnt := g.Count(); cnt != (N+N*N)-n {
			t.Fatalf("g.Count %d != %d", cnt, (N+N*N)-n)
		}
	}
	if n != N+N*N {
		t.Fatalf("Waited for %d != %d goroutines", n, N+N*N)
	}
	//t.Logf("n1 := %d, n2 = %d, total = %d", n1, n-n1, N+N*N)
}
