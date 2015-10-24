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
}

// GxcZero is a reset (zero) value for a Gcx goroutine context. See
// doc of Gcx.Go() method for its use.
var GcxZero Gcx

// ChKill is intented to be called from within the goroutine with
// context c, and returns the channel upon which the goroutine should
// wait for termination requests.
func (c *Gcx) ChKill() <-chan struct{} {
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
	go func(c *Gcx, f func() error) {
		err := f()
		c.status = err
		close(c.dead)
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

type Group struct {
	sync.Mutex
	n    int
	g    map[*Gcx]struct{}
	dead chan *Gcx
}
