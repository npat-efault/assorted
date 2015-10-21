// Package errors implements functions to manipulate errors. It can be
// used as a drop-in replacement to the very basic stdlib's "errors"
// package, and provides additional functionality, namelly: An error
// type that records location information (file-name,
// line-number). The ability to flag errors with characteristics such
// as "ErrTemporary" and "ErrTimeout"; errors thusly flagged can be
// checked using general predicate functions. The ability to "wrap"
// errors adding information to them and create error-stacks.
package errors

import "fmt"

// ShowLocations is a global configuration variable that controls
// whether error locations (file-name, line-number) are displayed. If
// "true", they are, if "false", they are not.
var ShowLocations bool = true

// Error flags used to signify error characteristics (e.g. "this is a
// temporary error") and help / guide the code handling the error. See
// also the ErrT type. Custom error types embedding ErrT can define
// additional flags if required, starting from 1 << ErrBitCustom. E.g:
//
//   const (
//       ErrAuthentication uint = 1 << ErrBitCustom + iota
//       ...
//   )
//
const (
	ErrTimeout uint = 1 << iota
	ErrTemporary

	ErrBitCustom = iota
)

// ErrT is a simple error type that you can use directly or embed in
// your own error types. It has a string message and a location
// (file-name, line-number) that can be optionally set (see functions
// Errf, Errf?NL). It can be flagged with characteristics like
// ErrTimeout and ErrTemporary. The presence of flags chan be checked
// using methods corresponding to the flag names (e.g. Err.Timeout()),
// or using the predicate functions like IsTimeout. Characteristics
// (flags) are used to help / guide the code handling the error. Types
// embedding Err can define additional flags.
type ErrT struct {
	Flags uint
	Loc   Location
	Msg   string
}

// Error formats ErrT as a string. Formating depends on the value of
// the global configuration flag ShowLocations.
func (e *ErrT) Error() string {
	if !ShowLocations || !e.Loc.IsSet() {
		return e.Msg
	}
	return e.Loc.String() + ": " + e.Msg
}

// Location returns ErrT's location. If no location is set for the
// error, then a zero-valued Location struct is returned.
func (e *ErrT) Location() Location {
	return e.Loc
}

// Timeout checks if Err has the ErrTimeout flag set
func (e *ErrT) Timeout() bool {
	return e.Flags&ErrTimeout != 0
}

// Temporary checks if Err has the ErrTemporary flag set
func (e *ErrT) Temporary() bool {
	return e.Flags&ErrTemporary != 0
}

// New creates and returns a new error. The error is not flagged with
// any characteristics, and its location is not set. New can be used
// as a drop-in replacement of stdlib's errors.New.
func New(msg string) error {
	return &ErrT{Msg: msg}
}

// Err creates and returns a new error. The error is flagged with
// "flags" (use the appropriate ErrXXX constants ORed together, or 0
// for no flags). The location of the error is set to the file-name and
// line-number of the Err invocation.
func Err(flags uint, msg string) error {
	e := &ErrT{Flags: flags, Msg: msg}
	e.Loc.Set(1)
	return e
}

// Errf creates and returns a new error using a Printf-like
// interface. See also function Err.
func Errf(flags uint, format string, a ...interface{}) error {
	e := &ErrT{Flags: flags, Msg: fmt.Sprintf(format, a...)}
	e.Loc.Set(1)
	return e
}

// ErrNL is similar with Err, with the difference that ErrNL does not
// set the error location. It can be used to create a global error
// value that can be returned from multiple source locations.
func ErrNL(flags uint, msg string) error {
	e := &ErrT{Flags: flags, Msg: msg}
	return e
}

// ErrfNL is similar with Err, with the difference that ErrNL does not
// set the error location. It can be used to create a global error
// value that can be returned from multiple source locations.
func ErrfNL(flags uint, format string, a ...interface{}) error {
	e := &ErrT{Flags: flags, Msg: fmt.Sprintf(format, a...)}
	return e
}

// IsTemporary is a predicate that tests if the error is a temporary
// one. It does so by checking if the concrete error type has a method
// with signature:
//
//    Temporary() bool
//
// and if that method returns "true" when called.
//
func IsTemporary(e error) bool {
	type tmpError interface {
		Temporary() bool
	}
	if et, ok := e.(tmpError); ok {
		return et.Temporary()
	}
	return false
}

// IsTimeout is a predicate that tests if the error indicates a
// Timeout. It does so by checking if the concrete error type has a
// method with signature:
//
//    Timeout() bool
//
// and if that method returns "true" when called.
//
func IsTimeout(e error) bool {
	type tmoError interface {
		Timeout() bool
	}
	if et, ok := e.(tmoError); ok {
		return et.Timeout()
	}
	return false
}
