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

	err = arithmetic.Encode(inFile, outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "got error while encoding: %v\n", err)
		return
	}

	inStat, err := inFile.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't obtain stat for input file: %v\n", err)
		return
	}

	outStat, err := outFile.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't obtain stat for output file: %v\n", err)
		return
	}

	inSize := inStat.Size()
	outSize := outStat.Size()
	ratio := float32(inStat.Size()) / float32(outStat.Size())
	log.Printf("input size: %v, output size: %v, ratio: %v\n", inSize, outSize, ratio)
}
