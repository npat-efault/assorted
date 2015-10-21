// Demonstrates a custom error type that embeds the basic error type
// "errors.ErrT". The custom type adds a custom field and defines a
// new custom flag.
package errors_test

import (
	"fmt"
	"strconv"

	"github.com/npat-efault/gohacks/errors"
)

// A custom Authorization error flag
const (
	ErrAuthorization uint = 1<<errors.ErrBitCustom + iota
)

// A custom error type with an additional int field
type customErr struct {
	errors.ErrT
	CustomField int
}

// Create and return a customErr with a given value (c) for the
// CustomField. Optionally flag the error with ErrTemporary and / or
// ErrAuthorization
func errCustom(c int, authflag bool, tmpflag bool) *customErr {
	ce := &customErr{}
	if authflag {
		ce.Flags |= ErrAuthorization
	}
	if tmpflag {
		ce.Flags |= errors.ErrTemporary
	}
	ce.Msg = "Custom Error"
	ce.Loc.Set(1)
	ce.CustomField = c
	return ce
}

// customErr's format customly
func (ce *customErr) Error() string {
	return ce.ErrT.Error() + " (" + strconv.Itoa(ce.CustomField) + ")"
}

// Test custom Authorization flag
func (ce *customErr) IsAuthorization() bool {
	return ce.Flags&ErrAuthorization != 0
}

// IsCustom error predicate
func IsAuthorization(e error) bool {
	type authErr interface {
		IsAuthorization() bool
	}
	if ec, ok := e.(authErr); ok {
		return ec.IsAuthorization()
	}
	return false
}

// Return a customErr #n (i.e. with CustomField == n)
func tst_custom(n int) error {
	return errCustom(n, false, false)
}

// Return a customErr #n, flagged with ErrAuthorization
func tst_custom_auth(n int) error {
	return errCustom(n, true, false)
}

// Return a customErr #n, flagged with ErrAuthorization and
// ErrTemporary
func tst_custom_auth_temp(n int) error {
	return errCustom(n, true, true)
}

func Example_custom() {
	// Enable display of error locations
	errors.ShowLocations = true
	// Display only base file names
	errors.LocationDisplay = errors.LocationBase

	if err := tst_custom(1); err != nil {
		// err must not be authorization, and not be temporary
		if !IsAuthorization(err) && !errors.IsTemporary(err) {
			fmt.Println(err)
		}
	}

	if err := tst_custom_auth(2); err != nil {
		// err must be authorization but not temporary
		if IsAuthorization(err) && !errors.IsTemporary(err) {
			fmt.Println("Auth:", err)
		}
	}

	if err := tst_custom_auth_temp(3); err != nil {
		// err must be authorization and temporary
		if IsAuthorization(err) && errors.IsTemporary(err) {
			fmt.Println("AuthTemp:", err)
		}
	}

	// Output:
	// custom_example_test.go:64: Custom Error (1)
	// Auth: custom_example_test.go:69: Custom Error (2)
	// AuthTemp: custom_example_test.go:75: Custom Error (3)
}
