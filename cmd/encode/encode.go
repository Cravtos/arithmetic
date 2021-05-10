package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/profile"

	"github.com/cravtos/arithmetic/internal/pkg/bitio"
	"github.com/cravtos/arithmetic/internal/pkg/config"
	"github.com/cravtos/arithmetic/internal/pkg/table"
)

// Interval delimiters
const top uint64 = (1 << config.IntervalBitsUsed) - 1
const firstQuart = (top + 1) / 4
const half = firstQuart * 2
const thirdQuart = firstQuart * 3

func bitsPlusFollow(to *bitio.Writer, bit uint64, bitsToFollow uint64) (err error) {
	if err = to.WriteBits(bit, 1); err != nil {
		return err
	}

	flipped := uint64(1)
	if bit == 1 {
		flipped = 0
	}

	for bitsToFollow > 0 {
		if err = to.WriteBits(flipped, 1); err != nil {
			return err
		}
		bitsToFollow--
	}

	return err
}

func main() {
	defer profile.Start(profile.ProfilePath("./profiling/encode/")).Stop()
	begin := time.Now()

	// Check if file is specified as argument
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v file_to_compress\n", os.Args[0])
		fmt.Println("Output is file_to_compress.arith")
		return
	}

	// Open file to read data
	inFilePath := filepath.Clean(os.Args[1])
	log.Println("opening file", inFilePath)
	inFile, err := os.Open(inFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't open file %s\n", inFilePath)
		return
	}
	defer inFile.Close()

	// Open file to write compressed data
	outFilePath := inFilePath + ".arith"
	log.Println("creating file", outFilePath)
	outFile, err := os.Create(outFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create file %s\n", outFilePath)
		return
	}
	defer outFile.Close()

	in := bufio.NewReader(inFile)
	out := bitio.NewWriter(outFile)
	t := table.NewTable()

	// Get information for header
	inStat, err := inFile.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't obtain stat for input file: %v\n", err)
		return
	}
	inSize := uint64(inStat.Size())

	// Write header
	log.Printf("writing header (number of encoded symbols): %v", inSize)
	if err := out.WriteBits(inSize, 64); err != nil {
		fmt.Fprintf(os.Stderr, "couldn't write header to output file: %v\n", err)
		return
	}

	// Begin compressing
	l := uint64(0)
	h := top
	bitsToFollow := uint64(0)

	log.Println("starting compression")
	v, err := in.ReadByte()
	nEncoded := 0
	for err == nil {
		denom := t.GetInterval(table.ABCSize - 1)
		delta := h - l + 1
		h = l + t.GetInterval(int(v))*delta/denom - 1
		l = l + t.GetInterval(int(v)-1)*delta/denom

		for {
			if h < half {
				if err := bitsPlusFollow(out, 0, bitsToFollow); err != nil {
					fmt.Fprintf(os.Stderr, "got error while writing bits to output file: %v\n", err)
					return
				}
				bitsToFollow = 0
			} else if l >= half {
				if err := bitsPlusFollow(out, 1, bitsToFollow); err != nil {
					fmt.Fprintf(os.Stderr, "got error while writing bits to output file: %v\n", err)
					return
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
				_, _ = fmt.Fprintf(os.Stderr, "got overflow\n")
				return
			}
		}

		t.UpdateCount(v)
		nEncoded += 1
		if nEncoded%config.UpdateRangesRate == 0 {
			t.UpdateRanges(0)
		}
		v, err = in.ReadByte()
	}

	// Encode last interval
	bitsToFollow += 1
	if l < firstQuart {
		if err := bitsPlusFollow(out, 0, bitsToFollow); err != nil {
			fmt.Fprintf(os.Stderr, "got error while writing bits to output file: %v\n", err)
			return
		}
	} else {
		if err := bitsPlusFollow(out, 1, bitsToFollow); err != nil {
			fmt.Fprintf(os.Stderr, "got error while writing bits to output file: %v\n", err)
			return
		}
	}
	bitsToFollow = 0

	// Write out full interval
	if err = out.WriteBits(l, config.IntervalBitsUsed); err != nil {
		fmt.Fprintf(os.Stderr, "got error while writing bits to output file: %v\n", err)
		return
	}

	// Flush everything to file
	if err := out.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "got error while flushing: %v\n", err)
		return
	}

	log.Println("finished. see", outFilePath)

	outStat, err := outFile.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't obtain stat for output file: %v\n", err)
		return
	}

	outSize := outStat.Size()
	ratio := float32(inStat.Size()) / float32(outStat.Size())
	log.Printf("input size: %v, output size: %v, ratio: %v\n", inSize, outSize, ratio)

	duration := time.Since(begin)
	log.Printf("time taken: %v\n", duration)
}
