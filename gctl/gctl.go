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
	ErrStarted    = errors.New("Goroutine already started")
	ErrNotStarted = errors.New("Goroutine has not been started")
)

// Gcx is a goroutine context. A pointer to a Gcx structure acts like
// a goroutine id: It identifies a specific goroutine and can be used
// to start it, kill it, and retrieve its exit status.
type Gcx struct {
	mu       sync.Mutex
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
// wait for a termination request. Termination is requested by closing
// the channel.
func (c *Gcx) ChKill() <-chan struct{} {
	c.mu.Lock()
	if c.kill == nil {
		c.mu.Unlock()
		panic("Gcx.ChKill: Gcx is not a running goroutine.")
	}
	c.mu.Unlock()
	return c.kill
}

// Go runs function f as a goroutine with context c. If a goroutine
// with context c has already been started, returns with
// ErrStarted. The goroutine terminates when function f returns. The
// return value of f is the goroutine's exit status, which can be
// retrieved using the method Gcx.Wait().
//
// Normally, once a goroutine with context c has run (and finished)
// you cannot start another one with the same context (as this can
// cause races, if you are not careful). Use a new context for each
// goroutine you start.
//
// CAVEAT: Nevertheless, if you want to reuse the same context, you
// can, provided that you first reset it by assigning to it the value
// GcxZero. In order to do so safely, you must be certain that the
// goroutine has indeed finished and that no-one will subcequently use
// the context to refer to the old goroutine. Again, it is easier not
// to reuse contexts, and in most cases there is no reason to.
func (c *Gcx) Go(f func() error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.kill != nil {
		return ErrStarted
	}
	c.kill = make(chan struct{})
	c.dead = make(chan struct{})
	if c.group != nil {
		c.group.mu.Lock()
		c.group.n++
		c.group.mu.Unlock()
	}
	go func(c *Gcx, f func() error) {
		err := f()
		c.status = err
		g := c.group
		// First close, then notify, in order to allow waiting
		// for an individual goroutine with Gcx.Wait, even if
		// it belongs to a group.
		close(c.dead)
		// Don't access c after this. Goroutine c is dead, and
		// they are allowed to zero-out c.
		if g != nil {
			// This may block until Group.Wait is called.
			g.notify <- c
		}
	}(c, f)
	return nil
}

// Kill signals goroutine with context c to stop. If the goroutine has
// already stopped, it does nothing. If no goroutine with context c
// has been started, it returns ErrNotStarted. It is ok to call Kill
// (for the same goroutine) concurrently from multiple goroutines.
func (c *Gcx) Kill() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.kill == nil {
		return ErrNotStarted
	}
	if c.signaled {
		return nil
	}
	c.signaled = true
	close(c.kill)
	return nil
}

// Wait waits for goroutine with context Gcx to finish, and returns
// the goroutine's exit status. If the goroutine has already finished,
// it returns imediatelly. If no goroutine with context c has been
// started, it returns ErrNotStarted. It is ok to call Wait (for the
// same goroutine) concurrently from multiple goroutines.
func (c *Gcx) Wait() error {
	c.mu.Lock()
	if c.kill == nil {
		c.mu.Unlock()
		return ErrNotStarted
	}
	c.mu.Unlock()
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
// wishes to wait on a number of goroutines and be notified when one
// (any) of them terminates.
//
// You can set the group of a goroutine by calling Gcx.SetGroup. A
// goroutine is considered member of the group from the time it is
// started (with Gcx.Go) until it terminates *and* a call to
// Group.Wait, Group.Poll or Group.Notify returns its
// exit-status. Getting the goroutine's exit status via Gcx.Wait does
// *not* remove it from the group.
//
// ATTENTION: If a group is garbage-collected (that is, all references
// to it are dropped) while it still has goroutines for which their
// exit status has not been retrieved (by calling Group.Wait,
// Group.Poll or Group.Notify) then you program will leak
// goroutines. Therefore, before dropping a group, make sure you have
// retrieved the exit status of all its member goroutines using
// Group.Wait, Group.Poll and/or Group.Notify.
type Group struct {
	mu     sync.Mutex
	n      int
	notify chan *Gcx
}

func (g *Group) init() {
	g.mu.Lock()
	if g.notify == nil {
		g.notify = make(chan *Gcx)
	}
	g.mu.Unlock()
}

// SetGroup sets the group of goroutine with context c to g. A
// goroutine can belong to only one group (or no group at all). The
// group of a goroutine must be set before it is started with
// Gcx.Go(). If SetGroup is called for an already started goroutine,
// it returns ErrStarted.
//
// Once added and started, the goroutine is considered member of the
// group. It remains so until it terminates *and* Group.Wait,
// Group.Poll, or Group.Notify return its exit status. See doc of
// Group.Wait for more.
func (c *Gcx) SetGroup(g *Group) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.kill != nil {
		return ErrStarted
	}
	g.init()
	c.group = g
	return nil
}

// Wait waits for one (any) of the goroutines in group g to
// terminate. It returns the context c of the goroutine that
// terminated, and its exit status. If upon entry to Group.Wait, the
// group has no more goroutines, it returns nil, nil.
//
// Once Group.Wait returns a goroutine's context and exit status, then
// the goroutine is no longer considered a member of the group.
func (g *Group) Wait() (c *Gcx, xs error) {
	g.mu.Lock()
	n := g.n
	g.mu.Unlock()
	if n == 0 {
		return nil, nil
	}
	c = <-g.notify
	g.mu.Lock()
	g.n--
	g.mu.Unlock()
	return c, c.Wait()
}

// Poll checks if a (any) goroutine in g has terminated, and if so
// returns its context and exit status. If upon entry to Group.Poll
// there is no goroutine that has terminated (or if the group is
// empty) it returns nil, nil.
//
// Once Group.Poll returns a goroutine's context and exit status, then
// the goroutine is no longer considered a member of the group.
func (g *Group) Poll() (c *Gcx, xs error) {
	g.mu.Lock()
	n := g.n
	g.mu.Unlock()
	if n == 0 {
		return nil, nil
	}
	select {
	case c = <-g.notify:
	default:
		return nil, nil
	}
	g.mu.Lock()
	g.n--
	g.mu.Unlock()
	return c, c.Wait()
}

// Count returns the number of goroutines in the group.
func (g *Group) Count() int {
	g.mu.Lock()
	n := g.n
	g.mu.Unlock()
	return n
}

// WaitAll calls Group.Wait repeatedly until all the goroutines in
// group g terminate. The exit status'es returned by the repeated
// calls to Group.Wait are discarded.
func (g *Group) WaitAll() {
	c, _ := g.Wait()
	for c != nil {
		c, _ = g.Wait()
	}
}

// ChNotify returns a channel upon which the caller can receive
// goroutine termination notifications. Each such notification is the
// context of a goroutine that has terminated. Once a notification is
// received the caller *MUST ALWAYS* call Group.Notify(), passing to
// it the received context. Group.Notify will return the goroutine's
// exit status and the goroutine will no longer be considered a member
// of the group.
//
// ChNotify / Notify are useful when one wishes to multiplex the wait
// for goroutine termination with other channel operations in a select
// statement.
func (g *Group) ChNotify() <-chan *Gcx {
	g.init()
	return g.notify
}

// Notify *MUST ALWAYS* and *ONLY* be called with the contexts
// received from the channel returned by Group.ChNotify. It returns
// the respective goroutine's exit status. *ANY* other use of Notify
// is an error and will leave the group in an invalid internal
// state. See also Group.ChNotify.
func (g *Group) Notify(c *Gcx) error {
	g.mu.Lock()
	g.n--
	g.mu.Unlock()
	return c.Wait()
}
