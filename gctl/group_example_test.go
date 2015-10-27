package gctl_test

import (
	"fmt"
	"time"

	"github.com/npat-efault/gohacks/gctl"
)

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
		return gctl.ErrKilled
	case <-time.After(5 * time.Duration(g.Num) * time.Millisecond):
		// Exit "normally" after 5 * g.Num milliseconds.
		// This way they finish in order of g.Num.
		return nil
	}
}

func ExampleGroup() {
	// Define a group for our goroutines
	grp := &gctl.Group{}
	// Keep track of the goroutines we start using a map. We use a
	// pointer to the goroutine's context (Gcx) as a key.
	gs := make(map[*gctl.Gcx]*G1)

	// Run 8 goroutines. Keep no more than 2 running at the same
	// time.
	for i := 0; i < 8; i++ {
		if grp.Count() == 2 {
			c, xs := grp.Wait()
			fmt.Printf("G%d exit status: %v\n", gs[c].Num, xs)
			delete(gs, c)
		}
		g := &G1{Num: i}
		g.SetGroup(grp)
		g.Go(g.run)
		gs[&g.Gcx] = g
	}

	// Wait for the rest to finish
	for c, xs := grp.Wait(); c != nil; c, xs = grp.Wait() {
		fmt.Printf("G%d exit status: %v\n", gs[c].Num, xs)
		// Goroutine has ended, remove it from our map
		delete(gs, c)
	}
	// Output:
	// G0 exit status: <nil>
	// G1 exit status: <nil>
	// G2 exit status: <nil>
	// G3 exit status: <nil>
	// G4 exit status: <nil>
	// G5 exit status: <nil>
	// G6 exit status: <nil>
	// G7 exit status: <nil>
}
