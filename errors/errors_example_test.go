// Demonstrates the basic usage of ther "errors" package.
package errors_test

import (
	"fmt"

	"github.com/npat-efault/gohacks/errors"
)

// You can define global error values like this. It's good not not to
// set the location for these global errors as they will be returned
// from multiple places in the source code. The advantage of
// pre-defining global errors rather than creating them on the spot
// (when you return them), is that you avoid allocating every time you
// return an error. The disadvantage is that you cannot (meaningfully)
// set the error location, and you cannot customize the message (or
// other params) per case.
var errGlobal = errors.New("Global error, no location")
var errGlobalTmo = errors.ErrNL(errors.ErrTimeout,
	"Global error, flagged as timeout")

// Returns an unflagged error. Error Location is set.
func tst_unflagged() error {
	return errors.Err(0, "Unflagged error with location")
}

// Returns an error flagged as ErrTemporay. Error location is set.
func tst_flagged() error {
	return errors.Err(errors.ErrTemporary,
		"Error flagged as temporary")
}

func Example() {
	// Enable display of error locations
	errors.ShowLocations = true
	// Display only base file names
	errors.LocationDisplay = errors.LocationBase

	if err := tst_unflagged(); err != nil {
		// err must have a set location
		if errors.Loc(err).IsSet() {
			fmt.Println(err)
		}
	}
	if err := tst_flagged(); err != nil {
		// err must be temporary and have a set location
		if errors.IsTemporary(err) && errors.Loc(err).IsSet() {
			fmt.Println(err)
		}
	}
	// errGlobal must not have a set Location
	if !errors.Loc(errGlobal).IsSet() {
		fmt.Println(errGlobal)
	}
	// errGlobalTmo must not have a set Location
	if !errors.Loc(errGlobalTmo).IsSet() {
		// errGlobalTmo must be flagged as timeout
		if errors.IsTimeout(errGlobalTmo) {
			fmt.Println(errGlobalTmo)
		}
	}
	// Output:
	// errors_example_test.go:24: Unflagged error with location
	// errors_example_test.go:30: Error flagged as temporary
	// Global error, no location
	// Global error, flagged as timeout
}
