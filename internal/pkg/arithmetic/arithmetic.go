package arithmetic

import (
	"bufio"
	"errors"
	"os"

	"github.com/cravtos/arithmetic/internal/pkg/config"
	"github.com/cravtos/arithmetic/internal/pkg/table"
	"github.com/icza/bitio"
)

// Interval delimiters
const top uint64 = (1 << config.IntervalBitsUsed) - 1
const firstQuart = (top + 1) / 4
const half = firstQuart * 2
const thirdQuart = firstQuart * 3

func bitsPlusFollow(to *bitio.Writer, bit uint64, bitsToFollow uint64) (err error) {
	to.TryWriteBitsUnsafe(bit, 1)

	flipped := uint64(1)
	if bit == 1 {
		flipped = 0
	}


	for bitsToFollow > 0 {
		to.TryWriteBitsUnsafe(flipped, 1)
		bitsToFollow--
	}

	return to.TryError
}

func Encode(inFile *os.File, outFile *os.File) (err error) {
	r := bufio.NewReader(inFile)
	w := bitio.NewWriter(outFile)

	t := table.NewTable()

	// Get information for header
	inStat, err := inFile.Stat()
	if err != nil {
		return err
	}
	inSize := uint64(inStat.Size())

	// Write header
	if err := w.WriteBitsUnsafe(inSize, 64); err != nil {
		return err
	}

	// Begin compressing
	l := uint64(0)
	h := top
	bitsToFollow := uint64(0)

	v, err := r.ReadByte()
	nEncoded := 0
	for err == nil {
		denom := t.GetInterval(table.ABCSize - 1)
		delta := h - l + 1
		h = l + t.GetInterval(int(v))*delta/denom - 1
		l = l + t.GetInterval(int(v)-1)*delta/denom

		for {
			if h < half {
				if err := bitsPlusFollow(w, 0, bitsToFollow); err != nil {
					return err
				}
				bitsToFollow = 0
			} else if l >= half {
				if err := bitsPlusFollow(w, 1, bitsToFollow); err != nil {
					return err
				}
				bitsToFollow = 0
				l -= half
				h -= half
			} else if l >= firstQuart && h < thirdQuart {
				bitsToFollow++
				l -= firstQuart
				h -= firstQuart
			} else {
				break
			}

			l <<= 1
			h <<= 1
			h += 1

			if l&top != l || h&top != h {
				return errors.New("got overflow")
			}
		}

		t.UpdateCount(v)
		nEncoded += 1
		if nEncoded%config.UpdateRangesRate == 0 {
			t.UpdateRanges(0)
		}
		v, err = r.ReadByte()
	}

	// Encode last interval
	bitsToFollow += 1
	if l < firstQuart {
		if err := bitsPlusFollow(w, 0, bitsToFollow); err != nil {
			return err
		}
	} else {
		if err := bitsPlusFollow(w, 1, bitsToFollow); err != nil {
			return err
		}
	}
	bitsToFollow = 0

	// Write full interval
	if err = w.WriteBits(l, config.IntervalBitsUsed); err != nil {
		return err
	}

	// Flush everything to file
	return w.Close()
}

func Decode(in *os.File, out *os.File) (err error) {
	r := bitio.NewReader(in)
	w := bufio.NewWriter(out)
	t := table.NewTable()

	// Read header
	var nEncoded uint64
	if nEncoded, err = r.ReadBits(64); err != nil {
		return err
	}

	// Begin decompressing
	value, err := r.ReadBits(config.IntervalBitsUsed)
	if err != nil {
		return err
	}

	l := uint64(0)
	h := top
	for i := uint64(1); i <= nEncoded; i++ {
		denom := t.GetInterval(table.ABCSize - 1)

		delta := h - l + 1
		interval := ((value-l+1)*denom - 1) / delta
		symbol := t.GetSymbol(interval)

		h = l + t.GetInterval(int(symbol))*delta/denom - 1
		l = l + t.GetInterval(int(symbol)-1)*delta/denom

		for {
			if h < half {
				// do nothing
			} else if l >= half {
				l -= half
				h -= half
				value -= half
			} else if l >= firstQuart && h < thirdQuart {
				l -= firstQuart
				h -= firstQuart
				value -= firstQuart
			} else {
				break
			}
			l <<= 1
			h <<= 1
			h += 1

			inBit, err := r.ReadBits(1)
			if err != nil {
				return err
			}
			value <<= 1
			value |= inBit & 1

			if l&top != l || h&top != h || value&top != value {
				return errors.New("got overflow")
			}
		}

		t.UpdateCount(symbol)
		if i%config.UpdateRangesRate == 0 {
			t.UpdateRanges(0)
		}
		if err = w.WriteByte(symbol); err != nil {
			return err
		}
	}

	return w.Flush()
}
