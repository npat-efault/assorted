package gctl_test

import (
	"fmt"
	"time"

	"github.com/npat-efault/gohacks/gctl"
)

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
		return gctl.ErrKilled
	case <-time.After(5 * time.Second):
		// Exit "normally" after 5 seconds
		return nil
	}
}

func ExampleGcx() {
	// Keep track of the gcx's we start using a slice
	gs := make([]*G, 0, 4)

	// Start 4 goroutines, each in its gcx
	for i := 0; i < 4; i++ {
		g := &G{Num: i} // g.Gcx is empty
		g.Go(g.run)     // g.Gcx is active
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
		// g.Gcx is dead
	}
	// Output:
	// Exiting G0
	// G0 status: Gcx context killed
	// Exiting G1
	// G1 status: Gcx context killed
	// Exiting G2
	// G2 status: Gcx context killed
	// Exiting G3
	// G3 status: Gcx context killed
}
