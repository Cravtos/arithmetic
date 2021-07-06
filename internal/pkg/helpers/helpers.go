package helpers

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// PrintRatio prints compression ratio for two files.
func PrintRatio(f *os.File, s *os.File) error {
	inStat, err := f.Stat()
	if err != nil {
		return err
	}

	outStat, err := s.Stat()
	if err != nil {
		return err
	}

	inSize := inStat.Size()
	outSize := outStat.Size()
	ratio := float32(inStat.Size()) / float32(outStat.Size())

	fmt.Printf("input size: %v\noutput size: %v\nratio: %v bytes\n", inSize, outSize, ratio)
	return nil
}

// CompareFiles returns true if two files are equal.
func CompareFiles(f *os.File, s *os.File) (bool, error) {
	const chunkSize = 64000

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f.Read(b1)
		if err1 != nil && err1 != io.EOF {
			return false, err1
		}

		b2 := make([]byte, chunkSize)
		_, err2 := s.Read(b2)
		if err2 != nil && err2 != io.EOF {
			return false, err2
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		} else if err1 == io.EOF || err2 == io.EOF {
			return false, nil
		}

		if !bytes.Equal(b1, b2) {
			return false, nil
		}
	}
}