package errors

import "fmt"

// WrappedSep is a configuration variable that defines the separator
// used when displaying wrapped errors with location information. By
// default wrapped errors are print on a new line, idented by a tab
// character.
var WrappedSep string = "\n\t"

type errWrap struct {
	msg string
	loc Location
	err error
}

func (e *errWrap) Error() string {
	if !ShowLocations {
		s := e.msg
		if e.err != nil {
			s += ": " + e.err.Error()
		}
		return s
	} else {
		s := fmt.Sprintf("%s: %s", e.loc, e.msg)
		if e.err != nil {
			if Loc(e.err).IsSet() {
				s += WrappedSep
			} else {
				s += ": "
			}
			s += e.err.Error()
		}
		return s
	}
}

func (e errWrap) wrappedError() error {
	return e.err
}

func (e errWrap) Location() Location {
	return e.loc
}

// Wrap returns an error that wraps the given error "e", adding a
// message and location information to it. The returned "wrapper"
// error can later itself be wrapped again, and again, creating
// something like a stack of errors.
func Wrap(e error, msg string) error {
	we := &errWrap{msg: msg, err: e}
	we.loc.Set(1)
	return we
}

// Wrapf works like Wrap, but has a Printf-like interface.
func Wrapf(e error, format string, a ...interface{}) error {
	we := &errWrap{msg: fmt.Sprintf(format, a...), err: e}
	we.loc.Set(1)
	return we
}

// Orig returns the original (bottom-most) error that is wrapped in a
// sequence of wrappers. If the error "e" is not a wrapper, then "e"
// itself is returned.
func Orig(e error) error {
	type errWrapper interface {
		wrappedError() error
	}
	for ew, ok := e.(errWrapper); ok; ew, ok = e.(errWrapper) {
		e = ew.wrappedError()
	}
	return e
}

// Wrapped returns the error that is wrapped by "e" (i.e. it "removes"
// the first wrapper). If "e" is not a wrapper, then it returns nil.
func Wrapped(e error) error {
	type errWrapper interface {
		wrappedError() error
	}
	if ew, ok := e.(errWrapper); ok {
		return ew.wrappedError()
	}
	return nil
}
