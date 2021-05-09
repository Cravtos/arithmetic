package bitio

import (
	"bufio"
	"errors"
	"io"
)

var (
	// ErrBigN occurs amount of bits to write is bigger than buffer.
	ErrBigN = errors.New("n is too big")
)

type Writer struct {
	to     *bufio.Writer
	buffer uint8
	bits   uint8
}

// NewWriter construct bitio.Writer from io.Writer.
func NewWriter(to io.Writer) *Writer {
	w := &Writer{
		to: bufio.NewWriter(to),
	}
	return w
}

// WriteBits writes n least significant bits from bv.
// Write won't happen if cumulative amount of bits in bv and in buffer is less than byte.
// Returns ErrBigN if n > 64 (greater than bv capacity).
func (w *Writer) WriteBits(bv uint64, n uint8) error {
	if n > 64 {
		return ErrBigN
	}

	// Null excess bits
	if n != 64 {
		bv = bv & ((1 << n) - 1)
	}

	newBits := w.bits + n

	// fill buffer, nothing to write
	if newBits < 8 {
		w.buffer |= uint8(bv) << (8 - newBits)
		w.bits = newBits
		return nil
	}

	if newBits > 8 {
		// fill buffer and write one byte
		write := 8 - w.bits
		err := w.to.WriteByte(w.buffer | uint8(bv>>(n-write)))
		if err != nil {
			return err
		}
		n -= write

		// write whole bytes
		for n >= 8 {
			n -= 8
			err = w.to.WriteByte(uint8(bv >> n))
			if err != nil {
				return err
			}
		}

		w.buffer = 0
		w.bits = 0

		// put remaining into cache
		if n > 0 {
			w.buffer = (uint8(bv) & ((1 << n) - 1)) << (8 - n)
			w.bits = n
		}
		return nil
	}

	// cache will be filled exactly with the bits to be written
	b := w.buffer | uint8(bv)
	w.buffer = 0
	w.bits = 0
	return w.to.WriteByte(b)
}

// WriteByte is alias for WriteBit(b, 8).
func (w *Writer) WriteByte(b byte) error {
	return w.WriteBits(uint64(b), 8)
}

// Flush writes buffer as it is and cleans it.
func (w *Writer) Flush() error {
	var err error

	if w.bits != 0 {
		if err = w.to.WriteByte(w.buffer); err != nil {
			return err
		}
	}

	w.buffer = 0
	w.bits = 0

	err = w.to.Flush()
	return err
}
