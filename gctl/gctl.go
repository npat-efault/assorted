// Copyright (c) 2015, Nick Patavalis (npat@efault.net).
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can
// be found in the LICENSE file.

// Package gctl provides simple goroutine control methods.
package gctl

import (
	"sync"

	"github.com/npat-efault/gohacks/errors"
)

var (
	ErrStarted    = errors.Errf(0, "Goroutine already started")
	ErrNotStarted = errors.Errf(0, "Goroutine has not been started")
)

// Gcx is a goroutine context. A pointer to a Gcx structure acts like
// a goroutine id: Identifies a specific goroutine and can be used
// to start it, kill it, and wait for its exit status.
type Gcx struct {
	sync.Mutex
	kill     chan struct{}
	dead     chan struct{}
	signaled bool
	status   error
	group    *Group
}

// GxcZero is the zero (reset) value for a Gcx goroutine context. See
// doc of Gcx.Go method for its use.
var GcxZero Gcx

// ChKill is intented to be called from within the goroutine with
// context c, and returns the channel upon which the goroutine should
// wait for termination requests.
func (c *Gcx) ChKill() <-chan struct{} {
	// Lock?
	if c.kill == nil {
		panic("Gcx.ChKill: Gcx is not a running goroutine.")
	}
	return c.kill
}

// Go runs function f as a goroutine with context c. If a goroutine
// with context c has already been started, returns with
// ErrStarted. The goroutine terminates when function f returns. The
// return value of f is the goroutine's exit status, which is returned
// by method Gcx.Wait(). Normally, once a goroutine with context c has
// run (and finished) you cannot start another one with the same
// context (as this can cause races, if you are not careful). If you
// want to do so, nevertheless, you must first reset the context by
// assigning to it the value GcxZero.
func (c *Gcx) Go(f func() error) error {
	c.Lock()
	defer c.Unlock()
	if c.kill != nil {
		return ErrStarted
	}
	c.kill = make(chan struct{})
	c.dead = make(chan struct{})
	if c.group != nil {
		c.group.Lock()
		c.group.n++
		c.group.Unlock()
	}
	go func(c *Gcx, f func() error) {
		err := f()
		c.status = err
		close(c.dead)
		if c.group != nil {
			c.group.notify <- c
		}
	}(c, f)
	return nil
}

// Kill signals goroutine with context c to stop. If the goroutine has
// already stopped, it does nothing. If no goroutine with context c
// has been started, it returns ErrNotStarted. It is ok to call Kill
// (for the same goroutine) concurrently from multiple goroutines.
func (c *Gcx) Kill() {
	c.Lock()
	defer c.Unlock()
	if c.kill == nil {
		return ErrNotStarted
	}
	if c.signaled {
		return
	}
	c.signaled = true
	close(c.kill)
}

// Wait waits for goroutine with context Gcx to finish, and returns
// the goroutine's exit status. If the goroutine has already finished,
// it returns imediatelly. If no goroutine with context c has been
// started, it returns ErrNotStarted. It is ok to call Wait (for the
// same goroutine) concurrently from multiple goroutines.
func (c *Gcx) Wait() error {
	// Lock?
	if c.kill == nil {
		return ErrNotStarted
	}
	<-c.dead
	return c.status
}

// KillWait signals goroutine with context Gcx to stop, then waits for
// it to finish, and returns its exit status. If the goroutine has
// already finished, it returns its exit status immediatelly. If no
// goroutine with context c has been started, it returns
// ErrNotStarted. It is ok to call KillWait (for the same goroutine)
// concurrently from multiple goroutines.
func (c *Gcx) KillWait() error {
	if err := c.Kill(); err != nil {
		return err
	}
	return c.Wait()
}

// Group groups together several goroutines. A group is used when one
// wishes to wait on a group of goroutines and be notified when one
// (any) of them terminates.
type Group struct {
	sync.Mutex
	n      int
	notify chan *Gcx
}

func (g *Group) init() {
	g.Lock()
	if g.notify == nil {
		g.notify = make(chan *Gcx)
	}
	g.Unlock()
}

// SetGroup adds goroutine with context c to group g. A goroutine can
// belong to only one group. A goroutine must be added to the group
// before it is started with Gcx.Go(). If SetGroup is called for an
// already running goroutine, it returns ErrStarted.
func (c *Gcx) SetGroup(g *Group) error {
	c.Lock()
	defer c.Unlock()
	if c.kill != nil {
		return ErrStarted
	}
	g.init()
	c.group = g
}

// Wait waits for one (any) of the goroutines in group g to
// terminate. It returns the context of the goroutine that terminated,
// and the number of running goroutines remaining in the group. The
// exit status of the goroutine that terminated can be acquired by
// subsequently calling the Wait method on the context returned;
// c.Wait(), in this case, will return immediatelly, since the
// goroutine has already terminated. If upon entry to Group.Wait the
// group has no running goroutines, it returns nil, 0
func (g *Group) Wait() (c *Gcx, n int) {
	g.Lock()
	n := g.n
	g.Unlock()
	if n == 0 {
		return nil, n
	}
	c := <-g.notify
	g.Lock()
	g.n--
	n = g.n
	g.Unlock()
	return c, n
}

// WaitAll waits for all the goroutines in group g to terminate.
func (g *Group) WaitAll() {
	_, n := g.Wait()
	for n != 0 {
		_, n := g.Wait()
	}
}
