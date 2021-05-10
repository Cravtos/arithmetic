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

func main() {
	defer profile.Start(profile.ProfilePath("./profiling/decode")).Stop()
	begin := time.Now()

	// Check if file is specified as argument
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v file_to_decompress\n", os.Args[0])
		fmt.Println("Output is file_to_compress.decoded")
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

	// Open file to write decompressed data
	outFilePath := inFilePath + ".decoded"
	log.Println("creating file", outFilePath)
	outFile, err := os.Create(outFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create file %s\n", outFilePath)
		return
	}
	defer outFile.Close()

	in := bitio.NewReader(inFile)
	out := bufio.NewWriter(outFile)
	t := table.NewTable()

	// Read header
	log.Printf("reading header (number of encoded symbols)")
	var nEncoded uint64
	if nEncoded, err = in.ReadBits(64); err != nil {
		fmt.Fprintf(os.Stderr, "couldn't read header from input file: %v\n", err)
		return
	}
	log.Printf("got number of encoded symbols: %v\n", nEncoded)

	// Begin decompressing
	log.Println("starting decompression")
	value, err := in.ReadBits(config.IntervalBitsUsed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "got error while reading input file (first %v bits): %v\n", config.IntervalBitsUsed, err)
		return
	}

	l := uint64(0)
	h := top
	for i := uint64(0); i < nEncoded; i++ {
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

			inBit, err := in.ReadBits(1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "got error while reading input file (on %v symbol): %v\n", i, err)
				return
			}
			value <<= 1
			value |= inBit & 1

			if l&top != l || h&top != h || value&top != value {
				_, _ = fmt.Fprintln(os.Stderr, "got overflow")
				return
			}
		}

		t.UpdateCount(symbol)
		if err = out.WriteByte(symbol); err != nil {
			fmt.Fprintf(os.Stderr, "got error while writting to output file: %v\n", err)
			return
		}
	}

	if err = out.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "got error while flushing output file: %v\n", err)
		return
	}

	log.Println("finished. see", outFilePath)

	// Get information for header
	inStat, err := inFile.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't obtain stat for input file: %v\n", err)
		return
	}
	inSize := uint64(inStat.Size())

	outStat, err := outFile.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't obtain stat for output file: %v\n", err)
		return
	}

	outSize := outStat.Size()
	ratio := float32(outStat.Size()) / float32(inStat.Size())
	log.Printf("input size: %v, output size: %v, ratio: %v\n", inSize, outSize, ratio)

	duration := time.Since(begin)
	log.Printf("time taken: %v\n", duration)
}
