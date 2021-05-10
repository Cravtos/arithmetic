package bitio

import (
	"bufio"
	"io"
)

type Reader struct {
	from   *bufio.Reader
	buffer uint8
	bits   uint8
}

// NewReader constructs new bitio.Reader from io.Reader.
func NewReader(from io.Reader) *Reader {
	r := &Reader{
		from: bufio.NewReader(from),
	}
	return r
}

// ReadBits reads n bits from Reader.
func (r *Reader) ReadBits(n uint8) (u uint64, err error) {
	// all bits are in buffer
	if n < r.bits {
		shift := r.bits - n
		u = uint64(r.buffer >> shift)
		r.buffer = r.buffer & (1<<shift - 1)
		r.bits = shift
		return u, nil
	}

	// need more bits than buffer has
	if n > r.bits {
		// take bits from buffer
		if r.bits > 0 {
			u = uint64(r.buffer)
			n -= r.bits
			r.bits = 0
		}

		// read whole bytes
		for n >= 8 {
			b, err := r.from.ReadByte()
			if err != nil {
				return 0, err
			}
			u = (u << 8) | uint64(b)
			n -= 8
		}

		// read last bits
		if n > 0 {
			r.buffer, err = r.from.ReadByte()
			if err != nil {
				return 0, err
			}
			shift := 8 - n
			u = (u << n) | uint64(r.buffer>>shift)
			r.buffer = r.buffer & (1<<shift - 1)
			r.bits = shift
		}

		return u, nil
	}

	r.bits = 0
	return uint64(r.buffer), nil
}

// ReadByte calls r.ReadBits(8).
func (r *Reader) ReadByte() (b byte, err error) {
	u, err := r.ReadBits(8)
	return byte(u), err
}
