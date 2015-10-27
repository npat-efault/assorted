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
	ErrGcxNotEmpty = errors.New("Gcx context not empty")
	ErrGcxEmpty    = errors.New("Gcx context is empty")
	ErrKilled      = errors.New("Gcx context killed")
)

// Gcx is a type that represents a goroutine context ("gcx", or
// "context"). A goroutine context is used to manage one or more
// related goroutines performing a certain task. A Gcx structure
// identifies a specific goroutine context and can be used to start
// it, kill it, wait for it to terminate, and retrieve its exit
// status.
//
// The zero value of the Gcx type is a valid goroutine context, ready
// to use. In this state the gcx is considered "empty".
//
// Method Gcx.Go starts a new goroutine in the context (and the
// context becomes "running" or "active"). Any goroutine in the
// context can call Gcx.Go to start more goroutines. The context
// terminates (becomes "dead") when all goroutines exit. No new
// goroutines can be started in a context that has terminated (is
// dead).
//
// Method Gcx.ChKill, called from within the context, returns a
// cancelation / termination channel that goroutines can monitor to
// receive termination requests. A single cancelation channel is used
// by all goroutines in the context.
//
// Method Gcx.Kill, called either from within, or from outside the
// context, closes the termination channel signaling the goroutines to
// exit.
//
// Method Gcx.Wait, called only from outside the context, waits until
// the context terminates, and returns its exit status.
type Gcx struct {
	mu       sync.Mutex
	kill     chan struct{} // close for termination request
	dead     chan struct{} // close when context dead
	ngort    int           // # of goroutines, -1: context dead
	signaled bool          // kill closed?
	status   error         // context exit status
	group    *Group
}

// GxcZero is the zero (empty) value for a Gcx goroutine context. See
// doc of Gcx.Go method for its use.
var GcxZero Gcx

// ChKill is intented to be called from the goroutines of context c,
// and returns the channel upon which the goroutines should wait for a
// termination / cancelation request. Termination is requested by
// closing the channel.
func (c *Gcx) ChKill() <-chan struct{} {
	c.mu.Lock()
	if c.kill == nil {
		c.mu.Unlock()
		panic("Gcx.ChKill: Gcx is not a running context")
	}
	c.mu.Unlock()
	return c.kill
}

// Go runs function f as a goroutine within context c. The goroutine
// terminates when function f returns. The return value of f is
// considered the goroutine's exit status. The context terminates when
// all it's goroutines terminate. The context's exit status is the
// first non-nil, non-ErrKilled exit status reported by its
// goroutines, or nil if all goroutines exited with nil, or ErrKilled
// if all goroutines exited with either nil, or ErrKilled. The
// context's exit status can be retrieved (the context can be
// "waited-for") using the method Gcx.Wait.
//
// A goroutine started with Gcx.Go can call Gcx.Go again to start
// additional goroutines in the same context.
//
// If a goroutine in c exits with a non-nil non-ErrKilled status, then
// the cancelation channel for c (Gcx.ChKill()) is closed, signaling
// all other goroutines in c to terminate.
//
// Normally, once a context c has run and terminated (its last
// goroutine has exited) it becomes "dead" and you cannot start it
// again. Calling Go on it after this point will panic.
//
// CAVEAT: Nevertheless, if you want to use the same Gcx context
// structure again, you can, provided that you first reset it by
// assigning to it the value GcxZero. In order to do so safely, you
// must make certain that the context is indeed dead and that no-one
// will subcequently use the same Gcx structure to logically refer to
// the old context. In any case, it is easier *not* to reuse context
// structures, and in most cases there is no reason to.
func (c *Gcx) Go(f func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ngort == -1 {
		panic("Gcx.Go: Gcx context is dead")
	}
	if c.kill == nil {
		c.kill = make(chan struct{})
		c.dead = make(chan struct{})
		if c.group != nil {
			c.group.mu.Lock()
			c.group.n++
			c.group.mu.Unlock()
		}
	}
	c.ngort++
	go func(c *Gcx, f func() error) {
		err := f()
		c.mu.Lock()
		if c.status == nil || c.status == ErrKilled {
			if err != nil {
				c.status = err
				if !c.signaled {
					close(c.kill)
					c.signaled = true
				}
			}
		}
		c.ngort--
		if c.ngort != 0 {
			c.mu.Unlock()
			return
		}

		// Last goroutine in context.
		c.ngort = -1 // mark as dead
		g := c.group
		c.mu.Unlock()

		// First close, then notify, in order to allow waiting
		// for an individual context with Gcx.Wait, even if it
		// belongs to a group.
		close(c.dead)
		// Don't access c after this. Context c is dead, and
		// they are allowed to zero-out c.
		if g != nil {
			// This may block until Group.Wait is
			// called.
			g.notify <- c
		}
	}(c, f)
}

// Kill signals goroutines in context c to stop by closing the channel
// returned by Gcx.ChKill. If the context is dead, it does nothing. If
// the context is empty, it returns ErrGcxEmpty. It is ok to call
// Kill from either within or outside the context. It is also ok to
// call Kill (for the same context) multiple times, or concurrently
// from multiple goroutines.
func (c *Gcx) Kill() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.kill == nil {
		return ErrGcxEmpty
	}
	if c.signaled {
		return nil
	}
	c.signaled = true
	close(c.kill)
	return nil
}

// Wait waits the context c to terminate (become dead), and returns
// the its exit status. If the context is already dead, it returns
// imediatelly. If the context is empty, it returns ErrGcxEmpty. It is
// ok to call Wait (for the same context) multiple times, or
// concurrently from multiple goroutines. Calling Wait from within
// context c leads to a deadlock.
func (c *Gcx) Wait() error {
	c.mu.Lock()
	if c.kill == nil {
		c.mu.Unlock()
		return ErrGcxEmpty
	}
	c.mu.Unlock()
	<-c.dead
	return c.status
}

// KillWait is the same as calling Gcx.Kill, followed by Gcx.Wait.
func (c *Gcx) KillWait() error {
	if err := c.Kill(); err != nil {
		return err
	}
	return c.Wait()
}

// Group groups together several gcx'es. A group is used when one
// wishes to wait on a number of contexts and be notified when one
// (any) of them terminates.
//
// You can set the group of a gcx by calling Gcx.SetGroup. A context
// is considered member of the group from the time it is started (it
// becomes running) until it terminates *and* a call to Group.Wait,
// Group.Poll or Group.Notify returns its exit-status. Getting the
// gcx's exit status via Gcx.Wait does *not* remove it from the
// group.
//
// ATTENTION: If a group is garbage-collected (that is, all references
// to it are dropped) while it still has gcx's for which their exit
// status has not been retrieved (by calling Group.Wait, Group.Poll or
// Group.Notify) then you program will leak goroutines. Therefore,
// before dropping a group, make sure you have retrieved the exit
// status of all its member gcx's using Group.Wait, Group.Poll and/or
// Group.Notify.
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

// SetGroup sets the group of gcx c to g. A gcx can belong to only one
// group (or no group at all). The group of a context must be set
// before it is started (befroe the first Gcx.Go call). If SetGroup is
// called for an already active gcx, it panics.
//
// Once added and started, the gcx is considered member of the
// group. It remains so until it terminates *and* Group.Wait,
// Group.Poll, or Group.Notify return its exit status. See doc of
// Group.Wait for more.
func (c *Gcx) SetGroup(g *Group) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.kill != nil {
		panic("Gcx.SetGroup: Gcx context not empty")
	}
	g.init()
	c.group = g
}

// Wait waits for one (any) of the contexts in group g to
// terminate. It returns a pointer to the Gcx structure of the context
// that terminated, and its exit status. If upon entry to Group.Wait,
// the group has no more gcx's, it returns nil, nil.
//
// Once Group.Wait returns a context and exit status, then the context
// is no longer considered a member of the group.
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

// Poll checks if a (any) gcx in g has already terminated, and if so
// returns a pointer to its Gcx structure and its exit status. If upon
// entry to Group.Poll there is no context that has terminated (or if
// the group is empty) it returns nil, nil.
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

// Count returns the number of gcx's in the group.
func (g *Group) Count() int {
	g.mu.Lock()
	n := g.n
	g.mu.Unlock()
	return n
}

// WaitAll calls Group.Wait repeatedly until all the gcx's in group g
// terminate. The exit status'es returned by the repeated calls to
// Group.Wait are discarded.
func (g *Group) WaitAll() {
	c, _ := g.Wait()
	for c != nil {
		c, _ = g.Wait()
	}
}

// ChNotify returns a channel upon which the caller can receive gcx
// termination notifications. Each such notification is a pointer to
// the Gcx structure of a context that has terminated. Once a
// notification is received the caller *MUST ALWAYS* call
// Group.Notify(), passing to it the received Gcx
// pointer. Group.Notify will return the gcx's exit status and the
// context will no longer be considered a member of the group.
//
// ChNotify / Notify are useful when one wishes to multiplex the wait
// for context termination with other channel operations in a select
// statement.
func (g *Group) ChNotify() <-chan *Gcx {
	g.init()
	return g.notify
}

// Notify *MUST ALWAYS* and *ONLY* be called with the context pointers
// received from the channel returned by Group.ChNotify. It returns
// the respective context's exit status. *ANY* other use of Notify is
// an error and will leave the group in an invalid internal state. See
// also Group.ChNotify.
func (g *Group) Notify(c *Gcx) error {
	g.mu.Lock()
	g.n--
	g.mu.Unlock()
	return c.Wait()
}
