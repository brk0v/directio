package directio

import (
	"errors"
	"io"
	"os"
	"unsafe"
)

const (
	// O_DIRECT alignment is 512B
	blockSize = 512

	// Default buffer is 8KB (2 pages).
	defaultBufSize = 8192
)

var _ io.WriteCloser = (*DirectIO)(nil)

// align returns an offset for alignment.
func align(b []byte) int {
	return int(uintptr(unsafe.Pointer(&b[0])) & uintptr(blockSize-1))
}

// allocAlignedBuf allocates buffer that is aligned by blockSize.
func allocAlignedBuf(n int) ([]byte, error) {
	if n == 0 {
		return nil, errors.New("size is `0` can't allocate buffer")
	}

	// Allocate memory buffer
	buf := make([]byte, n+blockSize)

	// First memmory alignment
	a1 := align(buf)
	offset := 0
	if a1 != 0 {
		offset = blockSize - a1
	}

	buf = buf[offset : offset+n]

	// Was alredy aligned. So just exit
	if a1 == 0 {
		return buf, nil
	}

	// Second alignment – check and exit
	a2 := align(buf)
	if a2 != 0 {
		return nil, errors.New("can't allocate aligned buffer")
	}

	return buf, nil
}

// DirectIO bypasses page cache.
type DirectIO struct {
	f   *os.File
	buf []byte
	n   int
	err error
}

// NewSize returns a new DirectIO writer.
func NewSize(f *os.File, size int) (*DirectIO, error) {
	if err := checkDirectIO(f.Fd()); err != nil {
		return nil, err
	}

	if size%blockSize != 0 {
		// align to blockSize
		size = size & -blockSize
	}

	if size < defaultBufSize {
		size = defaultBufSize
	}

	buf, err := allocAlignedBuf(size)
	if err != nil {
		return nil, err
	}

	return &DirectIO{
		buf: buf,
		f:   f,
	}, nil
}

// New returns a new DirectIO writer with default buffer size.
func New(f *os.File) (*DirectIO, error) {
	return NewSize(f, defaultBufSize)
}

// flush writes buffered data to the underlying os.File.
func (d *DirectIO) flush() error {
	if d.err != nil {
		return d.err
	}

	if d.n == 0 {
		return nil
	}

	n, err := d.f.Write(d.buf[0:d.n])

	if n < d.n && err == nil {
		err = io.ErrShortWrite
	}

	if err != nil {
		if n > 0 && n < d.n {
			copy(d.buf[0:d.n-n], d.buf[n:d.n])
		}
	}

	d.n -= n
	return err
}

// Flush writes buffered data to the underlying file.
func (d *DirectIO) Flush() error {
	fd := d.f.Fd()

	// Disable direct IO
	err := setDirectIO(fd, false)
	if err != nil {
		return err
	}

	// Making write without alignment
	err = d.flush()
	if err != nil {
		return err
	}

	// Enable direct IO back
	return setDirectIO(fd, true)
}

// Available returns how many bytes are unused in the buffer.
func (d *DirectIO) Available() int { return len(d.buf) - d.n }

// Buffered returns the number of bytes that have been written into the current buffer.
func (d *DirectIO) Buffered() int { return d.n }

// Write writes the contents of p into the buffer.
// It returns the number of bytes written.
// If nn < len(p), it also returns an error explaining
// why the write is short.
func (d *DirectIO) Write(p []byte) (nn int, err error) {
	// Write more than available in buffer.
	for len(p) >= d.Available() && d.err == nil {
		var n int
		// Check if buffer is zero size for direct and zero copy write to Writer.
		// Here we also check the p memory alignment.
		// If buffer p is not aligned, than write through buffer d.buf and flush.
		if d.Buffered() == 0 && align(p) == 0 {
			// Large write, empty buffer.
			if (len(p) % blockSize) == 0 {
				// Data and buffer p are already aligned to block size.
				// So write directly from p to avoid copy.
				n, d.err = d.f.Write(p)
			} else {
				// Data needs alignment. Buffer alredy aligned.

				// Align data
				l := len(p) & -blockSize

				// Write directly from p to avoid copy.
				var nl int
				nl, d.err = d.f.Write(p[:l])

				// Save other data to buffer.
				n = copy(d.buf[d.n:], p[l:])
				d.n += n

				// written and buffered data
				n += nl
			}
		} else {
			n = copy(d.buf[d.n:], p)
			d.n += n
			err = d.flush()
			if err != nil {
				return nn, err
			}
		}
		nn += n
		p = p[n:]
	}

	if d.err != nil {
		return nn, d.err
	}

	n := copy(d.buf[d.n:], p)
	d.n += n
	nn += n

	return nn, nil
}

func (d *DirectIO) Close() error {
	if d.err == nil {
		err := d.Flush()
		if err != nil {
			return err
		}
	}
	return d.f.Close()
}
