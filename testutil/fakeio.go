package testutil

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"
)

var (
	ErrClosed    = errors.New("Closed")
	ErrTemporary = errors.New("Temporary Error")
	ErrPermanent = errors.New("Permanent Error")
)

// FakeIO if a buffer that provides io.ReadCloser and io.WriteCloser
// interfaces, crudely simulating a connection with delays and
// errors. The FakeIO buffer is, generally, *not* thread safe; you
// cannot issue Read and Write operations on the same buffer
// concurently from multiple goroutines. However, you *can* call Close
// concurently (i.e. from a different goroutine) with ongoing Read or
// Write operations.
type FakeIO struct {
	limit    int
	errAfter int
	errEvery int
	delay    time.Duration
	countR   int
	countW   int
	closed   chan struct{}
	buff     bytes.Buffer
}

// NewFakeIO initializes and returns a new FakeIO buffer. The buffer
// returned is ready for Read and Write operations. If the buffer is
// to be used for Read operations, it, most likely, must first be
// "filled" by calling FakeIO.FillString(), FakeIO.FillBytes(), or
// FakeIO.FillFile(). The arguments "limit", "errAfter", "errEvery",
// and "delay" control the behavior of the buffer in Read an Write
// operations. Specifically: Argument "limit" controls the maximum
// amount of bytes that can be read from the buffer with a single Read
// call. Argument "errAfter" is the number of Read or Write calls
// after which all subsequent Read or Write calls will fail with
// ErrPermanent. Argument "errEvery" causes the buffer to fail every
// "errEvery" Read or Write call with ErrTemporary (e.g if "errEvery"
// == 2, the 2nd, 4th, 6th, etc. calls will fail). Read and Write
// calls are counted separately towards "errAfter" and
// "errEvery". Argument "delay" causes Read and Write operations to
// delay for the specified amount before they return (either
// succesfully or with error).
func NewFakeIO(limit, errAfter, errEvery int, delay time.Duration) *FakeIO {
	f := &FakeIO{}
	f.limit = limit
	f.errAfter = errAfter
	f.errEvery = errEvery
	f.delay = delay
	f.buff.Reset()
	f.countR, f.countW = 0, 0
	f.closed = make(chan struct{})
	return f
}

// FillString fills the yet unread part of the buffer with data from
// the string.
func (f *FakeIO) FillString(s string) {
	f.buff.WriteString(s)
}

// FillBytes fills the yet unread part of the buffer with data from
// the byte-slice.
func (f *FakeIO) FillBytes(b []byte) {
	f.buff.Write(b)
}

// FillFile fills the yet unread part of the buffer with data from
// the named file.
func (f *FakeIO) FillFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(&f.buff, file)
	return err
}

// Bytes returns the yet-unread data in the buffer.
func (f *FakeIO) Bytes() []byte {
	return f.buff.Bytes()
}

// Close closes the FakeIO buffer. Ongoing and subsequent Read, Write,
// and Close calls will fail with ErrClosed. After Close, in order to
// reuse the buffer, you must call Init (and possibly FillXXX)
// again. It *is* ok to call Close from a different goroutine,
// concurently with ongoing Read and Write operations on the same
// buffer.
func (f *FakeIO) Close() error {
	select {
	case <-f.closed:
		return ErrClosed
	default:
		close(f.closed)
		return nil
	}
}

func (f *FakeIO) Read(p []byte) (n int, err error) {
	select {
	case <-f.closed:
		return 0, ErrClosed
	default:
	}
	f.countR++
	if f.delay != 0 {
		select {
		case <-time.After(f.delay):
		case <-f.closed:
			return 0, ErrClosed
		}
	}
	if f.buff.Len() == 0 {
		return 0, io.EOF
	}
	if f.errAfter != 0 && f.countR > f.errAfter {
		return 0, ErrPermanent
	}
	if f.errEvery != 0 && f.countR%f.errEvery == 0 {
		return 0, ErrTemporary
	}
	if f.limit != 0 && len(p) > f.limit {
		p = p[:f.limit]
	}

	return f.buff.Read(p)
}

func (f *FakeIO) Write(p []byte) (n int, err error) {
	select {
	case <-f.closed:
		return 0, ErrClosed
	default:
	}
	f.countW++
	if f.delay != 0 {
		select {
		case <-time.After(f.delay):
		case <-f.closed:
			return 0, ErrClosed
		}
	}
	if f.errAfter != 0 && f.countW > f.errAfter {
		return 0, ErrPermanent
	}
	if f.errEvery != 0 && f.countW%f.errEvery == 0 {
		return 0, ErrTemporary
	}

	return f.buff.Write(p)
}
