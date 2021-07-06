package main

import (
	"fmt"
	"github.com/cravtos/arithmetic/internal/pkg/arithmetic"
	"log"
	"os"
	"path/filepath"
)



func main() {

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

	err = arithmetic.Decode(inFile, outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "got error while decoding: %v\n", err)
		return
	}

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
}
