package gctl_test

import (
	"fmt"
	"time"

	"github.com/npat-efault/gohacks/errors"
	"github.com/npat-efault/gohacks/gctl"
)

var ErrKilled = errors.New("Killed")

type G struct {
	gctl.Gcx // Embed goroutine context
	// ... any additional fields
	Num int
}

// Goroutine entry point
func (g *G) run() error {
	select {
	case <-g.ChKill():
		// Exit with ErrKilled, if signaled
		fmt.Printf("Exiting G%d\n", g.Num)
		return ErrKilled
	case <-time.After(5 * time.Second):
		// Exit "normally" after 5 seconds
		return nil
	}
}

func ExampleGcx() {
	// Keep track of the goroutines we start using a slice
	gs := make([]*G, 0, 4)

	// Start 4 goroutines
	for i := 0; i < 4; i++ {
		g := &G{Num: i}
		if err := g.Go(g.run); err != nil {
			// Can't really happen
			fmt.Printf("Failed to start G%d\n", g.Num)
		}
		// Add the goroutine we started to the slice
		gs = append(gs, g)
	}

	// Kill and Wait them
	for i := 0; i < 4; i++ {
		g := gs[i]
		err := g.KillWait()
		if err != nil {
			fmt.Printf("G%d status: %v\n", g.Num, err)
		}
	}
	// Output:
	// Exiting G0
	// G0 status: Killed
	// Exiting G1
	// G1 status: Killed
	// Exiting G2
	// G2 status: Killed
	// Exiting G3
	// G3 status: Killed
}
