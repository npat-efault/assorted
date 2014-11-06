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
// errors. The fields "Limit", "ErrAfter", "ErrEvery", and "Delay" are
// option fields that control the behavior of the buffer for Read an
// Write operations. The FakeIO buffer is, generally, *not* thread
// safe; you *cannot* issue Read and Write operations on the same
// buffer, or modify the option fields, concurently, from multiple
// goroutines. However, you *can* call Close concurently (i.e. from a
// different goroutine) with ongoing Read or Write operations.
type FakeIO struct {
	// Max number of bytes that can be read with a single
	// call. Zero means no limit.
	Limit int
	// Number of Read/Write calls after which all subsequent
	// Read/Write calls will fail with ErrPermanent. Read and
	// Write calls are counted separatelly. Zero means never.
	ErrAfter int
	// Cause every ErrEvery Read / Write call to fail with
	// ErrTemporary (e.g. if ErrEvery == 2, the 2nd, 4th, 6th,
	// etc. calls will fail). Read and Write calls are counted
	// separatelly. Zero means never.
	ErrEvery int
	// Delay read and write operationes for the specified
	// amount. Zero means no delay.
	Delay  time.Duration
	countR int
	countW int
	closed chan struct{}
	buff   bytes.Buffer
}

// NewFakeIO initializes and returns a new FakeIO buffer. All option
// fields are set to their zero value. The buffer returned is ready
// for Read and Write operations. If the buffer is to be used for Read
// operations, it, most likely, must first be "filled" by calling
// FakeIO.FillString(), FakeIO.FillBytes(), or FakeIO.FillFile().
func NewFakeIO() *FakeIO {
	f := &FakeIO{}
	f.closed = make(chan struct{})
	return f
}

// Reset empties the buffer and prepares it for Read and Write
// operations (even if it was closed). Reset does not affect the
// option fields.
func (f *FakeIO) Reset() {
	f.buff.Reset()
	f.countR, f.countW = 0, 0
	f.closed = make(chan struct{})
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
// reuse the buffer, you must call Reset (and possibly FillXXX)
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
	if f.Delay != 0 {
		select {
		case <-time.After(f.Delay):
		case <-f.closed:
			return 0, ErrClosed
		}
	}
	if f.buff.Len() == 0 {
		return 0, io.EOF
	}
	if f.ErrAfter != 0 && f.countR > f.ErrAfter {
		return 0, ErrPermanent
	}
	if f.ErrEvery != 0 && f.countR%f.ErrEvery == 0 {
		return 0, ErrTemporary
	}
	if f.Limit != 0 && len(p) > f.Limit {
		p = p[:f.Limit]
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
	if f.Delay != 0 {
		select {
		case <-time.After(f.Delay):
		case <-f.closed:
			return 0, ErrClosed
		}
	}
	if f.ErrAfter != 0 && f.countW > f.ErrAfter {
		return 0, ErrPermanent
	}
	if f.ErrEvery != 0 && f.countW%f.ErrEvery == 0 {
		return 0, ErrTemporary
	}

	return f.buff.Write(p)
}
