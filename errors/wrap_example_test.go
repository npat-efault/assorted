// Demonstrates error wrapping
package errors_test

import (
	"fmt"

	"github.com/npat-efault/gohacks/errors"
)

func baz() error {
	return errors.Err(0, "Cannot do baz")
}

func bar() error {
	if err := baz(); err != nil {
		return errors.Wrap(err, "baz failed")
	}
	return nil
}

func foo() error {
	if err := bar(); err != nil {
		return errors.Wrap(err, "bar failed")
	}
	return nil
}

func Example_wrap() {
	// Enable display of error locations
	errors.ShowLocations = true
	// Display package-name and base file-name
	errors.LocationDisplay = errors.LocationPackage

	if err := foo(); err != nil {
		// Show complete error stack
		fmt.Printf("Failed: %s\n", err)
		// Show only original error
		fmt.Printf("Failed: %s", errors.Orig(err))
	}
	// Output:
	// Failed: errors/wrap_example_test.go:23: bar failed
	// 	errors/wrap_example_test.go:16: baz failed
	// 	errors/wrap_example_test.go:11: Cannot do baz
	// Failed: errors/wrap_example_test.go:11: Cannot do baz
}
