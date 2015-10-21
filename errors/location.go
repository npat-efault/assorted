package errors

import (
	"fmt"
	"path"
	"runtime"
)

// LocationDisplayMode is a type that encodes the available modes
// (formats) for displaying error locations. See LocationXXX constants
// for valid values.
type LocationDisplayMode int

// Location display modes
const (
	LocationFull    LocationDisplayMode = iota // Full path name
	LocationPackage                            // Pkg/file-name
	LocationBase                               // Base file-name
)

// LocationDisplay is a global configuration variable that controls
// the way error locations are displayed. See LocationXXX constants
// for valid values.
var LocationDisplay LocationDisplayMode = LocationPackage

func trimFile(f string) string {
	switch LocationDisplay {
	case LocationPackage:
		return path.Base(path.Dir(f)) + "/" + path.Base(f)
	case LocationBase:
		return path.Base(f)
	default:
		return f
	}
}

// Location is a type encoding error locations (as file-name and
// line-number pairs).
type Location struct {
	File string
	Line int
}

// IsSet method tests if the error location is set
func (l Location) IsSet() bool {
	return l.File != ""
}

// String method returns the error location formated as a string. The
// way the location is formated depends on the value of the
// LocationDisplay global variable.
func (l Location) String() string {
	if !l.IsSet() {
		return ""
	}
	return fmt.Sprintf("%s:%d", trimFile(l.File), l.Line)
}

// Set sets the location to the position where the method was called
// from. The "skip" argument indicates the number of stack-frames to
// skip when setting the location. If 0, the location is set to the
// line of the Set() method invocation. If 1, the location is set to
// the line of the invocation of the function that contains the Set
// method invocation... and so on, all the way up the call stack.
//
// Assuming this code (numbers on the left are line numbers):
//
//   10:  var l errors.Location
//   11:  func foo() { bar() }
//   12:  func bar() { baz() }
//   13:  func baz() { qux() }
//   14:  func qux() { l.Set(skip) }
//
// Location is set like this:
//
//   skip = 0, l.Line = 14
//   skip = 1, l.Line = 13
//   skip = 2, l.Line = 12
//   skip = 3, l.Line = 11
//
func (l *Location) Set(skip int) {
	_, l.File, l.Line, _ = runtime.Caller(skip + 1)
}

// Loc returns the location of the error "e". This function can be
// used with any error type. If the type does not have a location
// record, or if it does, but the the location is not set, then a
// zero-valued Location structure is returned.
func Loc(e error) Location {
	type errWithLocation interface {
		Location() Location
	}
	if el, ok := e.(errWithLocation); ok {
		return el.Location()
	}
	return Location{}
}
