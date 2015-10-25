package gctl_test

import (
	"fmt"
	"time"

	"github.com/npat-efault/gohacks/errors"
	"github.com/npat-efault/gohacks/gctl"
)

var ErrKilled1 = errors.New("Killed")

type G1 struct {
	gctl.Gcx // Embed goroutine context
	// ... any additional fields
	Num int
}

// Goroutine entry point
func (g *G1) run() error {
	select {
	case <-g.ChKill():
		// Exit with ErrKilled, if signaled
		fmt.Printf("Exiting G%d\n", g.Num)
		return ErrKilled1
	case <-time.After(5 * time.Duration(g.Num) * time.Millisecond):
		// Exit "normally" after 5 * g.Num milliseconds.
		// This way they finish in order of g.Num.
		return nil
	}
}

// Define a group for our goroutines
var Grp = &gctl.Group{}

func ExampleGroup() {
	// Keep track of the goroutines we start using a map. We use a
	// pointer to the goroutine's context (Gcx) as a goroutine
	// identifier.
	gs := make(map[*gctl.Gcx]*G1)

	// Start 4 goroutines
	for i := 0; i < 4; i++ {
		g := &G1{Num: i}
		g.SetGroup(Grp)
		if err := g.Go(g.run); err != nil {
			// Won't really happen
			fmt.Printf("Failed to start G%d\n", g.Num)
		}
		// Add the goroutine we started to our map.
		gs[&g.Gcx] = g
	}

	// Wait them to finish normally, in any order, and report
	// their exit status.
	for c, xs := Grp.Wait(); c != nil; c, xs = Grp.Wait() {
		fmt.Printf("G%d exit status: %v\n", gs[c].Num, xs)
		// Goroutine has ended, remove it from our map
		delete(gs, c)
	}
	// Output:
	// G0 exit status: <nil>
	// G1 exit status: <nil>
	// G2 exit status: <nil>
	// G3 exit status: <nil>
}
